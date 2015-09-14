package server

import (
	"log"
	"net"
	"net/rpc"
 	"net/http"
	"os"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
	// "github.com/rjkroege/leap/search"
)


type Server struct {
}

type QueryBundle struct {
	Fn string
	Stype string
	Suffix string
}

type QueryResult struct {
	Entries []output.Entry
}

func BeginServing(config *base.Configuration)  {
	// May wish to place self in new process group.	

	state := &Server{}
	rpc.Register(state)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}

// Need to parse args myself.
func (t *Server) Leap(query QueryBundle, resultBuffer *QueryResult) error {
	log.Println("go leap remoted: ", query)
	*resultBuffer = QueryResult{[]output.Entry{}}
	return nil
}


func (t *Server) Shutdown(ignored string, result *string) error {
	log.Println("shutting down...")
	os.Exit(0)
	return nil
}
