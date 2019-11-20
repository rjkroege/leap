package client

import (
	"log"
	"net/rpc"

	"github.com/rjkroege/leap/server"
)

// RpcRequester implements BlockSourceRequester over the Go rpc channel
// to the leap server instance running on the remote.
type RpcRequester struct {
	leapserver *rpc.Client
	// Add additional state.

	token int
}

func MakeRpcRequester(s *rpc.Client, token int) *RpcRequester {
	return &RpcRequester{
		leapserver: s,
		token:      token,
	}
}

// Copied from HttpRequester
func (r *RpcRequester) IsFatal(err error) bool {
	return true
}

// DoRequest is executed by the core of the rsync code to fetch a desired
// block of the remote file. It's synchronous (BlockSourceBase takes care
// of dispatching possibly many of these requests.)
func (r *RpcRequester) DoRequest(startOffset int64, endOffset int64) ([]byte, error) {
	log.Println("DoRequest got asked")
	req := server.DoRequestArgs{
		Start: startOffset,
		End:   endOffset,
		Token: r.token,
	}
	var buffy []byte
	err := r.leapserver.Call("Server.DoRequestOnServer", req, &buffy)

	return buffy, err
}
