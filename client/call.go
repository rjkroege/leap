package client

import (
	"net/rpc"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
	// "github.com/rjkroege/leap/search"
	"github.com/rjkroege/leap/server"
)

func RemoteInvokeQuery(_ *base.Configuration, query server.QueryBundle) ([]output.Entry, error) {
	// TODO(rjk): fix up the addressing and pass in config
	serverAddress := "localhost"
	client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		return nil, err
	}

	// Synchronous call
	// TODO(rjk): need to create some kind of argument packetization.
	var reply server.QueryResult
	err = client.Call("Server.Leap", query, &reply)
	if err != nil {
		return nil, err
	}
	
	return reply.Entries, nil
}
