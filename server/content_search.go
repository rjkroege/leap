package server

import (
	"fmt"

	"github.com/rjkroege/leap/output"
)

type ContentSearchResultArgs struct {
	Fnames []int32
	Suffix string
}

type ContentSearchResult struct {
	Entries []output.Entry
}

func (s *Server) RemoteContentSearchResult(args ContentSearchResultArgs, resp *ContentSearchResult) error {
	return fmt.Errorf("not implemented yets")
}
