package search

import (
	"encoding/xml"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/rjkroege/leap/output"
)

func testIndex(t *testing.T) string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("can't figure out the path to this test file")
	}
	return filepath.Join(filepath.Dir(thisFile), "test_index")
}

func tDir(rpath string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), rpath)
}

func TestGetTestDataPath(t *testing.T) {
	if got, expected := testIndex(t), "/Users/rjkroege/tools/gopkg/src/github.com/rjkroege/leap/search/test_index"; got != expected {
		t.Errorf("got %#v expected %#v", got, expected)
	}
}

func TestNoMatchFileNameOnlyQuery(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	// No files have z in them.
	expected := []output.Entry{}
	got, err := gen.Query(".*z.*", "", "")
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestOneMatchFileNameOnlyQuery(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	// One file has c in the name.
	expected := []output.Entry{output.Entry{XMLName: xml.Name{Space: "",
		Local: ""},
		Uid:          tDir("test_data/b/ccc.txt/good bye\n"),
		Arg:          tDir("test_data/b/ccc.txt:1"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "ccc.txt:1 good bye\n",
		SubTitle:     "b/ccc.txt:1 good bye\n",
		Icon: output.AlfredIcon{Filename: tDir("test_data/b/ccc.txt"),
			Type: "fileicon"}}}

	got, err := gen.Query(".*c.*", "", "")
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}