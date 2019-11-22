package server

import (
	"fmt"
)

type DoRequestArgs struct {
	Start, End int64
	Token      int
}

// DoRequestOnServer runs on the server and returns the requested blocks.
func (t *Server) DoRequestOnServer(req DoRequestArgs, resp *[]byte) error {
	if t.token == 0 || t.token != req.Token {
		return fmt.Errorf("token mis-match: new sync before last one is done")
	}

	e := req.End
	s := req.Start

	buffy := make([]byte, e-s)
	if _, err := t.indexfile.ReadAt(buffy, s); err != nil {
		return err
	}

	// TODO(rjk): compress the blocks here.

	*resp = buffy
	return nil
}
