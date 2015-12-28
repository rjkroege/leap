package server

import (
	"log"
	"net"
	"net/rpc"
 	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
)


type Server struct {
	search output.Generator
	ftime time.Time
	config *base.Configuration
	lock sync.Mutex
}

type QueryBundle struct {
	Fn []string
	Stype string
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

func BeginServing(config *base.Configuration)  {
	// Stash date of the index file that we actually use.
	ftime := getFileTime(config.Indexpath)

	// Need to take index path from Configuration.
	state := &Server{ search: search.NewTrigramSearch(config.Indexpath, config.Prefixes), ftime: ftime, config: config }
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

	log.Println("go leap remoted: ", query)
	entries, err := t.search.Query(query.Fn, query.Stype, []string{query.Suffix})
	log.Println(entries)
	*resultBuffer = QueryResult{entries}
	return err
}

func (t *Server) Shutdown(ignored string, result *string) error {
	log.Println("shutting down...")
	os.Exit(0)
	return nil
}
