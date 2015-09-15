package client

import (
	"net/rpc"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
	// "github.com/rjkroege/leap/search"
	"github.com/rjkroege/leap/server"
)

// TODO(rjk): Make the port configurable?
func RemoteInvokeQuery(config *base.Configuration, query server.QueryBundle) ([]output.Entry, error) {
	serverAddress := config.Hostname
	client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		return nil, err
	}

	// Synchronous call
	var reply server.QueryResult
	err = client.Call("Server.Leap", query, &reply)
	if err != nil {
		return nil, err
	}
	
	return reply.Entries, nil
}

func Shutdown(config *base.Configuration) error {
	serverAddress := config.Hostname
	client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		return err
	}

	var reply string
	err = client.Call("Server.Shutdown", "", &reply)
	if err != nil {
		return err
	}
	return nil
}
