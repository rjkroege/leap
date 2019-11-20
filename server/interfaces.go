package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/Redundancy/go-sync/chunks"
	"github.com/Redundancy/go-sync/filechecksum"
	grsync "github.com/Redundancy/go-sync/index"
	"github.com/Redundancy/go-sync/indexbuilder"
	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/index"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
)

// Configuration is for mocking the Configuration code.
// TODO(rjk): This was expedient but not very nice. I should consider
// cleaning this up later.
type Configuration interface {
	GetNewConfiguration() *base.GlobalConfiguration
	ClassicConfiguration() *base.Configuration
}

type ReaderAtCloser interface {
	io.ReaderAt
	io.Closer
}

type Server struct {
	search    output.Generator
	ftime     time.Time
	config    Configuration
	lock      sync.Mutex
	indexfile ReaderAtCloser

	// It's conceivable that I don't need token?
	token int

	indexer Indexer
	fs      filesystem
	build   builder
}

type QueryBundle struct {
	Fn     []string
	Stype  string
	Suffix string
}

type QueryResult struct {
	Entries []output.Entry
}

func getFileTime(filename string) time.Time {
	finfo, err := os.Stat(filename)
	if err != nil {
		log.Fatal("couldn't open the index file: ", err)
	}
	return finfo.ModTime()
}

// TODO(rjk): This setup assumes that the remote index path is valid.
// It might not be. I need to both handle the situation that the remote
// has no index file and/or that it fails. Perhaps BeginServing needs to be
// written for testability.
func BeginServing(config Configuration) {
	// Stash date of the index file that we actually use.
	ftime := getFileTime(config.ClassicConfiguration().Indexpath)

	// Need to take index path from Configuration.
	state := &Server{search: search.NewTrigramSearch(config.ClassicConfiguration().Indexpath, config.ClassicConfiguration().Prefixes), ftime: ftime, config: config, indexer: index.Idx{}, fs: filesystemimpl{}, build: builderimpl{}}

	// The argument to rpc.Register can be any interface. It's public methods become the
	// methods available on the server via Go rpc.
	rpc.Register(state)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}

func (t *Server) checkTimeAndUpdate() {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Compare the date with the stashed one.
	ctime := getFileTime(t.config.ClassicConfiguration().Indexpath)
	if t.ftime.Before(ctime) {
		// Must reload the index file here.
		// TODO(rjk): this needs to be in a lock?
		t.ftime = ctime
		t.search = search.NewTrigramSearch(t.config.ClassicConfiguration().Indexpath, t.config.ClassicConfiguration().Prefixes)
	}
}

// Need to parse args myself.
func (t *Server) Leap(query QueryBundle, resultBuffer *QueryResult) error {
	t.checkTimeAndUpdate()

	entries, err := t.search.Query(query.Fn, query.Stype, []string{query.Suffix})
	*resultBuffer = QueryResult{entries}
	return err
}

func (t *Server) Shutdown(_ string, result *string) error {
	log.Println("shutting down...")
	os.Exit(0)
	return nil
}

const (
	// I made this up. I have no idea if its reasonable.
	// I suppose that I can benchmark.
	KILOBYTE   = 1024
	BLOCK_SIZE = 4 * KILOBYTE
)

type IndexAndBuildChecksumIndexArgs struct {
	Token             int
	RemoteProjectName string
	RemotePath        string
}

type RemoteCheckSumIndexData struct {
	CindexOutput         []byte
	FileSize             int64
	ReferenceFileIndex   *grsync.ChecksumIndex
	StrongChecksumGetter chunks.StrongChecksumGetter
}

// Indexer lets me mock out the interface to the use of cindex.
type Indexer interface {
	ReIndex(indexpath string, args ...string) ([]byte, error)
}

// filesystem permits replacing os.Stat.
type filesystem interface {
	Stat(string) (os.FileInfo, error)
}

type filesystemimpl struct{}

func (_ filesystemimpl) Stat(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// builder permits replacing BuildChecksumIndex
type builder interface {
	BuildChecksumIndex(*filechecksum.FileChecksumGenerator, io.Reader) (
		[]byte,
		*grsync.ChecksumIndex,
		filechecksum.ChecksumLookup,
		error,
	)
}

type builderimpl struct{}

func (_ builderimpl) BuildChecksumIndex(check *filechecksum.FileChecksumGenerator, r io.Reader) (
	[]byte,
	*grsync.ChecksumIndex,
	filechecksum.ChecksumLookup,
	error,
) {
	return indexbuilder.BuildChecksumIndex(check, r)
}

func (s *Server) IndexAndBuildChecksumIndex(args IndexAndBuildChecksumIndexArgs, resp *RemoteCheckSumIndexData) error {
	if s.token != 0 && s.token != args.Token {
		return fmt.Errorf("token mis-match: two syncs in progress?")
	}
	s.token = args.Token

	// Get configuration.
	newconfig := s.config.GetNewConfiguration()
	if newconfig == nil {
		s.token = 0
		return fmt.Errorf("index command requires upgrading config")
	}

	// Re-index..
	stdout, err := s.indexer.ReIndex(args.RemotePath)
	if err != nil {
		s.token = 0
		return fmt.Errorf("remote index command failed because: %v", err)
	}
	resp.CindexOutput = stdout

	// NB: would be easy to do this for all the cases. (In a later CL)
	// stat here to get the file size
	indexpath := args.RemotePath
	fileinfo, err := s.fs.Stat(indexpath)
	if err != nil {
		s.token = 0
		return fmt.Errorf("can't stat remote index %s because %v", indexpath, err)
	}
	resp.FileSize = fileinfo.Size()

	// Open and stash the open file
	indexfile, err := os.Open(indexpath)
	if err != nil {
		s.token = 0
		return fmt.Errorf("can't open remote index %s because %v", indexpath, err)
	}
	s.indexfile = indexfile

	generator := filechecksum.NewFileChecksumGenerator(BLOCK_SIZE)
	_, referenceFileIndex, checksumLookup, err := s.build.BuildChecksumIndex(generator, indexfile)
	if err != nil {
		s.token = 0
		indexfile.Close()
		return fmt.Errorf("can't compute checksums on %s because %v", indexpath, err)
	}

	// Shove the files types into the response.
	resp.ReferenceFileIndex = referenceFileIndex
	scg, ok := checksumLookup.(chunks.StrongChecksumGetter)
	if !ok {
		log.Println("I deeply misunderstand how the sync code works")
		indexfile.Close()
		s.token = 0
		return fmt.Errorf("can't convert checksumLookup into a concrete StrongChecksumGetter")
	}
	resp.StrongChecksumGetter = scg

	return nil
}
