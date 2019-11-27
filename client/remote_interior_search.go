package client

import (
	"fmt"
	"net/rpc"

	"github.com/google/codesearch/regexp"
	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/server"
)

type RemoteInternalSearcher struct {
	prefixes    []string
	remoteindex string
	leapserver  *rpc.Client
}

func (ris *RemoteInternalSearcher) ContentSearchResult(fnames []uint32, re *regexp.Regexp, suffix string) ([]output.Entry, error) {
	args := server.ContentSearchResultArgs{
		Fnames:      fnames,
		Suffix:      suffix,
		Prefixes:    ris.prefixes,
		Remoteindex: ris.remoteindex,
	}
	var reply server.ContentSearchResult

	if err := ris.leapserver.Call("Server.RemoteContentSearchResult", args, &reply); err != nil {
		return nil, fmt.Errorf("can't invoke RemoteContentSearchResult on server: %v", err)
	}
	return reply.Entries, nil
}

func NewRemoteInternalSearcher(config *base.Configuration) (*RemoteInternalSearcher, error) {
	newconfig := config.GetNewConfiguration()
	localproject := newconfig.Currentproject
	serverAddress := newconfig.Projects[localproject].Host

	leapserver, err := rpc.DialHTTP("tcp", serverAddress+":1234")
	if err != nil {
		return nil, fmt.Errorf("can't connect to %s:1234: %v", serverAddress, err)
	}

	return &RemoteInternalSearcher{
		prefixes:    newconfig.Projects[localproject].Prefixes,
		remoteindex: newconfig.Projects[localproject].Remotepath,
		leapserver:  leapserver,
	}, nil
}
