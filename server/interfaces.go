package server

import (
	"log"
	"net"
	"net/rpc"
 	"net/http"
	"os"

	"github.com/rjkroege/leap/base"
	// "github.com/rjkroege/leap/output"
	// "github.com/rjkroege/leap/search"
)


type Server struct {
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

func (t *Server) Leap(query string, resultBuffer *string) error {
	log.Println("go leap remoted: ", query)
	*resultBuffer = "hello there"
	return nil
}


func (t *Server) Shutdown(ignored string, result *string) error {
	log.Println("shutting down...")
	os.Exit(0)
	return nil
}
