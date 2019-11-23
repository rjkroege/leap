package client

import (
	"fmt"

	"github.com/rjkroege/leap/base"
	"github.com/google/codesearch/regexp"
	"github.com/rjkroege/leap/output"
)

type RemoteInternalSearcher struct {
}

func (ris* RemoteInternalSearcher) ContentSearchResult(fnames []uint32, re *regexp.Regexp, suffix string) ([]output.Entry, error) {
		return nil, fmt.Errorf("not implemented")
}

func NewRemoteInternalSearcher(_ *base.Configuration) (*RemoteInternalSearcher, error) {
	return nil, fmt.Errorf("not implemented")
}
