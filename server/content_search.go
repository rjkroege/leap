package server

import (
	"fmt"

	"github.com/google/codesearch/regexp"
	"github.com/rjkroege/leap/output"
)

type ContentSearchResultArgs struct {
	Fnames      []uint32
	Suffix      string
	Prefixes    []string
	Remoteindex string
}

type ContentSearchResult struct {
	Entries []output.Entry
}

func (s *Server) RemoteContentSearchResult(args ContentSearchResultArgs, resp *ContentSearchResult) error {
	// Make sure that we are using the most recent index data. We do this
	// here instead of making it part of the index implementation because I
	// might have run cindex.
	if err := s.ensureValidSearchObject(args.Remoteindex, args.Prefixes); err != nil {
		return fmt.Errorf("server can't make search object for %s: %v", args.Remoteindex, err)
	}

	re, err := regexp.Compile(args.Suffix)
	if err != nil {
		return fmt.Errorf("can't compile regexp on server: %v", err)
	}
	entries, err := s.search.ContentSearchResult(args.Fnames, re, "")
	if err != nil {
		return fmt.Errorf("can't run Search.ContentSearchResult on server: %v", err)
	}

	resp.Entries = entries
	return nil
}
