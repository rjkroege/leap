package server

import (
	"fmt"
	"testing"
)

type TestReaderAt struct {
	contents string
	err      error
}

func (ra *TestReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	copy(p, ra.contents[off:])
	return len(p), ra.err
}

const buffercontents = "hello there i am a buffer"

func TestDoRequestOnServer(t *testing.T) {

	file := &TestReaderAt{
		contents: buffercontents,
		err:      nil,
	}

	server := &Server{
		indexfile: file,
	}

	var result []byte
	if err := server.DoRequestOnServer(DoRequestArgs{
		Start: 0,
		End:   5,
		Token: 1,
	}, &result); err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if want, got := "hello", string(result); want != got {
		t.Errorf("unexpected response: got %s want %s\n", got, want)
	}

	var result2 []byte
	if err := server.DoRequestOnServer(DoRequestArgs{
		Start: int64(len(buffercontents) - 6),
		End:   int64(len(buffercontents)),
		Token: 1,
	}, &result2); err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if want, got := "buffer", string(result2); want != got {
		t.Errorf("unexpected response: got %s want %s\n", got, want)
	}

	brokenfile := &TestReaderAt{
		contents: buffercontents,
		err:      fmt.Errorf("fake failure"),
	}
	server.indexfile = brokenfile
	var result3 []byte
	if err := server.DoRequestOnServer(DoRequestArgs{
		Start: int64(len(buffercontents) - 6),
		End:   int64(len(buffercontents)),
		Token: 1,
	}, &result3); err == nil {
		t.Errorf("expected error")
	} else {
		if want, got := brokenfile.err.Error(), err.Error(); want != got {
			t.Errorf("unexpected response: got %s want %s\n", got, want)
		}
	}
}
