package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
	"io"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/index"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
	"github.com/Redundancy/go-sync/filechecksum"
	"github.com/Redundancy/go-sync/chunks"
	grsync "github.com/Redundancy/go-sync/index"
	"github.com/Redundancy/go-sync/indexbuilder"
)

type Server struct {
	search output.Generator
	ftime  time.Time
	config *base.Configuration
	lock   sync.Mutex
	indexfile io.ReaderAt

	// It's conceivable that I don't need token?
	token int
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

func BeginServing(config *base.Configuration) {
	log.Println("BeginServing: ", *config, *config.GetNewConfiguration())
	// Stash date of the index file that we actually use.
	ftime := getFileTime(config.Indexpath)

	// Need to take index path from Configuration.
	state := &Server{search: search.NewTrigramSearch(config.Indexpath, config.Prefixes), ftime: ftime, config: config}

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
	ctime := getFileTime(t.config.Indexpath)
	if t.ftime.Before(ctime) {
		// Must reload the index file here.
		// TODO(rjk): this needs to be in a lock?
		t.ftime = ctime
		t.search = search.NewTrigramSearch(t.config.Indexpath, t.config.Prefixes)
	}
}

// Need to parse args myself.
func (t *Server) Leap(query QueryBundle, resultBuffer *QueryResult) error {
	t.checkTimeAndUpdate()

	entries, err := t.search.Query(query.Fn, query.Stype, []string{query.Suffix})
	*resultBuffer = QueryResult{entries}
	return err
}

func (t *Server) Shutdown(ignored string, result *string) error {
	log.Println("shutting down...")
	os.Exit(0)
	return nil
}

// TODO(rjk): deprecated. Remove this code.
// Index implemens the remote Index command. The client needs to
// specify the remote project name.
func (t *Server) Index(remoteprojectname string, result *string) error {
	newconfig := t.config.GetNewConfiguration()
	if newconfig == nil {
		return fmt.Errorf("index command requires upgrading config")
	}

	// TODO(rjk): Capture the output and send it back?
	stdout, err := index.ReIndex(newconfig, remoteprojectname)
	*result = string(stdout)
	return err
}


const (
	// I made this up. I have no idea if its reasonable.
	// I suppose that I can benchmark.
	KILOBYTE = 1024 
	BLOCK_SIZE = 4 * KILOBYTE
)


type IndexAndBuildChecksumIndexArgs struct {
	Token int
	RemoteProjectName string
	RemotePath string
}

type RemoteCheckSumIndexData struct {
	CindexOutput []byte
	FileSize int64
	ReferenceFileIndex *grsync.ChecksumIndex
	StrongChecksumGetter  chunks.StrongChecksumGetter
}

func (s *Server) IndexAndBuildChecksumIndex(args IndexAndBuildChecksumIndexArgs, resp *RemoteCheckSumIndexData) error {
	if s.token != 0 && s.token != args.Token {
		return fmt.Errorf("Token mis-match: two syncs in progress?")
	}
	s.token = args.Token

	// Get configuration.
	newconfig := s.config.GetNewConfiguration()
	if newconfig == nil {
		return fmt.Errorf("index command requires upgrading config")
	}

	// Re-index..
	stdout, err := index.ReIndex(newconfig, args.RemoteProjectName)
	if err != nil {
		return fmt.Errorf("remote index command failed because: %v", err)
	}
	resp.CindexOutput = stdout
	
	// NB: would be easy to do this for all the cases. (In a later CL)
	// stat here to get the file size
	indexpath := args.RemotePath
	fileinfo, err := os.Stat(indexpath)
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
	_, referenceFileIndex, checksumLookup, err := indexbuilder.BuildChecksumIndex(generator, indexfile)
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
		return fmt.Errorf("can't convert checksumLookup into a concrete StrongChecksumGetter")
	}
	resp.StrongChecksumGetter = scg

	return nil
}

type DoRequestArgs struct {
	Start, End int64
	Token int
}

// DoRequestOnServer runs on the server and returns the requested blocks.
func (t *Server) DoRequestOnServer(req DoRequestArgs, resp *[]byte) error {
	// TODO(rjk): Server needs to contain a indexfile os.ReaderAt 
	// TODO(rjk): validate the provided token here.

	e := req.End
	s := req.Start

	buffy := make([]byte, e - s)
	if _, err := t.indexfile.ReadAt(buffy, s); err != nil {
		return err
	}

	// TODO(rjk): compress the blocks here.

	*resp = buffy
	return nil
}
