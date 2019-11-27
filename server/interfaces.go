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
	ftime  time.Time
	config Configuration
	lock   sync.Mutex
	search *search.Search

	indexfile ReaderAtCloser
	token     int

	indexer Indexer
	fs      filesystem
	build   builder
}

func getFileTime(filename string) (time.Time, error) {
	finfo, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, fmt.Errorf("couldn't Stat index file: %v", err)
	}
	return finfo.ModTime(), nil
}

// Prefixes are used in the server for trimming as part of the
// implementation of ContentSearchResult. It is conceivable that this is
// undesirable. It is the case that the prefixes can be passed in as part
// of each RemoteContentSearchResult RPC invocation. There is no need to
// persist them as part of the Server object.
func BeginServing(config Configuration) {
	state := &Server{
		// Do I needz config?
		config:  config,
		indexer: index.Idx{},
		fs:      filesystemimpl{},
		build:   builderimpl{},
	}

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

func (t *Server) ensureValidSearchObject(indexname string, prefixes []string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Always get the time of the possibly new indexfile.
	ntime, err := getFileTime(indexname)
	if err != nil {
		return fmt.Errorf("can't stat open indexfile %s: %v", indexname, err)
	}

	if t.search != nil && t.search.GetName() == indexname && !t.ftime.Before(ntime) {
		return nil
	}

	// I have to make a new Search instance. Clean up the old one.
	if t.indexfile != nil {
		t.indexfile.Close()
	}
	// TODO(rjk): Cleanup. The lack of cleanup here will cause the server
	// side to leak memory.
	t.search = search.NewTrigramSearch(indexname, prefixes)
	t.ftime = ntime

	return nil
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
	RemotePath string
}

type RemoteCheckSumIndexData struct {
	Token                int
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
	// Re-index
	stdout, err := s.indexer.ReIndex(args.RemotePath)
	if err != nil {
		return fmt.Errorf("remote index command failed because: %v", err)
	}
	resp.CindexOutput = stdout

	// NB: would be easy to do this for all the cases. (In a later CL)
	// stat here to get the file size
	indexpath := args.RemotePath
	fileinfo, err := s.fs.Stat(indexpath)
	if err != nil {
		return fmt.Errorf("can't stat remote index %s because %v", indexpath, err)
	}
	resp.FileSize = fileinfo.Size()

	// Open and stash the open file
	indexfile, err := os.Open(indexpath)
	if err != nil {
		return fmt.Errorf("can't open remote index %s because %v", indexpath, err)
	}
	s.indexfile = indexfile

	generator := filechecksum.NewFileChecksumGenerator(BLOCK_SIZE)
	_, referenceFileIndex, checksumLookup, err := s.build.BuildChecksumIndex(generator, indexfile)
	if err != nil {
		indexfile.Close()
		return fmt.Errorf("can't compute checksums on %s because %v", indexpath, err)
	}

	// Shove the files types into the response.
	resp.ReferenceFileIndex = referenceFileIndex
	scg, ok := checksumLookup.(chunks.StrongChecksumGetter)
	if !ok {
		log.Println("I deeply misunderstand how the sync code works")
		indexfile.Close()
		return fmt.Errorf("can't convert checksumLookup into a concrete StrongChecksumGetter")
	}
	resp.StrongChecksumGetter = scg
	s.token += 1
	resp.Token = s.token
	return nil
}
