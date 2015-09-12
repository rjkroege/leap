package client

import (
	"net/rpc"

	// "github.com/rjkroege/leap/base"
	// "github.com/rjkroege/leap/output"
	// "github.com/rjkroege/leap/search"
)

func RemoteInvokeQuery(query string) (string, error) {
	// TODO(rjk): fix up the addressing and pass in config
	serverAddress := "localhost"
	client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		return "", err
	}

	// Synchronous call
	// TODO(rjk): need to create some kind of argument packetization.
	var reply string
	err = client.Call("Server.Leap", "call from client", &reply)
	if err != nil {
		return "", err
	}
	
	return reply, nil
}
