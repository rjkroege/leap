package server

type DoRequestArgs struct {
	Start, End int64
	Token int
}

// DoRequestOnServer runs on the server and returns the requested blocks.
func (t *Server) DoRequestOnServer(req DoRequestArgs, resp *[]byte) error {
	// TODO(rjk): Server needs to contain a indexfile os.ReaderAt 
	// TODO(rjk): validate the provided token here.

	e := req.End
	s := req.Start

	buffy := make([]byte, e - s)
	if _, err := t.indexfile.ReadAt(buffy, s); err != nil {
		return err
	}

	// TODO(rjk): compress the blocks here.

	*resp = buffy
	return nil
}
