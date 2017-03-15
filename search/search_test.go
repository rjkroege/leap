package search

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
)

func testIndex(t *testing.T) string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("can't figure out the path to this test file")
	}
	return filepath.Join(filepath.Dir(thisFile), "test_index")
}

func tDir(rpath ...string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(append([]string{filepath.Dir(thisFile)}, rpath...)...)
}

func pDir(num int, rpath ...string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return "/" + filepath.Join(append([]string{fmt.Sprintf("/%s:%d", base.Prefix, num), filepath.Dir(thisFile)}, rpath...)...)
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
	got, err := gen.Query([]string{".*z.*"}, "", []string{""})
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
		Uid:          tDir("test_data/b/ccc.txt"),
		Arg:          tDir("test_data/b/ccc.txt"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "ccc.txt",
		SubTitle:     "b/ccc.txt",
		Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"},
	}}

	got, err := gen.Query([]string{".*c.*"}, ":", []string{""})
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestOneMatchFileNameLineNumberQuery(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	// One file has c in the name.
	expected := []output.Entry{output.Entry{XMLName: xml.Name{Space: "",
		Local: ""},
		Uid:          tDir("test_data/b/ccc.txt"),
		Arg:          tDir("test_data/b/ccc.txt:2"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "ccc.txt:2",
		SubTitle:     "b/ccc.txt:2",
		Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"}}}

	got, err := gen.Query([]string{".*c.*"}, ":", []string{"2"})
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestOneMatchContentQuery(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	// One file contains carrot.
	expected := []output.Entry{output.Entry{XMLName: xml.Name{Space: "",
		Local: ""},
		Uid:          tDir("test_data/b/aaa.txt:2"),
		Arg:          pDir(2, "test_data/b/aaa.txt"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "2 carrot\n",
		SubTitle:     ".../aaa.txt:2 carrot\n",
		Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"}}}

	got, err := gen.Query([]string{""}, "/", []string{"carrot"})
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestBadFileRegexp(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	// One file contains carrot.
	_, err := gen.Query([]string{")*"}, "/", []string{""})
	if err == nil {
		t.Errorf("unexpected absence of error on query: %v\n", err)
	}
}

func TestBadContentRegexp(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	// One file contains carrot.
	_, err := gen.Query([]string{""}, "/", []string{")*"})
	if err == nil {
		t.Errorf("unexpected absence of error on query: %v\n", err)
	}
}

func TestOneMatchFileNameLineNumberQueryWithPrefix(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), []string{
		tDir(""),
	})

	// One file has c in the name.
	expected := []output.Entry{output.Entry{XMLName: xml.Name{Space: "",
		Local: ""},
		Uid:          tDir("test_data/b/ccc.txt"),
		Arg:          tDir("test_data/b/ccc.txt:2"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "ccc.txt:2",
		SubTitle:     "test_data/b/ccc.txt:2",
		Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"}}}

	got, err := gen.Query([]string{".*c.*"}, ":", []string{"2"})
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestOneMatchFileNameLineNumberQueryWithSlashedPrefix(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), []string{
		tDir("") + "/",
	})

	// One file has c in the name.
	expected := []output.Entry{output.Entry{XMLName: xml.Name{Space: "",
		Local: ""},
		Uid:          tDir("test_data/b/ccc.txt"),
		Arg:          tDir("test_data/b/ccc.txt:2"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "ccc.txt:2",
		SubTitle:     "test_data/b/ccc.txt:2",
		Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"}}}

	got, err := gen.Query([]string{".*c.*"}, ":", []string{"2"})
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestMissingFile(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	os.Rename("test_data/b/aaa.txt", "test_data/b/aaa.txt.missing")
	defer os.Rename("test_data/b/aaa.txt.missing", "test_data/b/aaa.txt")

	expected := []output.Entry{output.Entry{XMLName: xml.Name{Space: "",
		Local: ""},
		Uid:          tDir("test_data/b/ccc.txt:4"),
		Arg:          pDir(4, "test_data/b/ccc.txt"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "4 beet\n",
		SubTitle:     ".../ccc.txt:4 beet\n",
		Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"}}}

	// Inject log
	txtlog := new(bytes.Buffer)
	log.SetOutput(txtlog)

	// Run query
	got, err := gen.Query([]string{""}, "/", []string{"beet"})

	// Put the log back.
	log.SetOutput(os.Stderr)

	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
	lgot := txtlog.String()
	lexpected := "multiFile error:  open " + tDir("test_data/b/aaa.txt") + ": no such file or directory"
	if strings.Index(lgot, lexpected) == -1 {
		t.Errorf("got %#v exepcted %#v", lgot, lexpected)
	}
}

func TestLargeFile(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	// One file contains carrot.
	expected := []output.Entry{output.Entry{XMLName: xml.Name{Space: "",
		Local: ""},
		Uid:          tDir("test_data/b/bbb.txt:7617"),
		Arg:          pDir(7617, "test_data/b/bbb.txt"),
		Type:         "file",
		Valid:        "",
		AutoComplete: "",
		Title:        "7617 turnip",
		SubTitle:     ".../bbb.txt:7617 turnip",
		Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"}}}

	got, err := gen.Query([]string{""}, "/", []string{"turnip"})

	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestManyMatchesFile(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	expected := make([]output.Entry, MaximumMatches)
	for i, _ := range expected {
		num := fmt.Sprintf("%d", i+1)
		expected[i] = output.Entry{XMLName: xml.Name{Space: "",
			Local: ""},
			Uid:          tDir("test_data/b/ddd.txt:" + num),
			Arg:          pDir(i+1, "test_data/b/ddd.txt"),
			Type:         "file",
			Valid:        "",
			AutoComplete: "",
			Title:        num + " broccoli\n",
			SubTitle:     ".../ddd.txt:" + num + " broccoli\n",
			Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"}}
	}

	got, err := gen.Query([]string{""}, "/", []string{"broccoli"})

	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestMultiRegexpFileNameOnlyQuery(t *testing.T) {
	gen := NewTrigramSearch(testIndex(t), nil)

	expected := make([]output.Entry, 0)
	for _, fn := range []string{"bbb.txt", "aaa.txt", "ccc.txt", "ddd.txt"} {
		expected = append(expected, output.Entry{
			XMLName:      xml.Name{Space: "", Local: ""},
			Uid:          tDir("test_data/b", fn),
			Arg:          tDir("test_data/b", fn),
			Type:         "file",
			Valid:        "",
			AutoComplete: "",
			Title:        fn,
			SubTitle:     "b/" + fn,
			Icon:         output.AlfredIcon{Filename: "/Applications/TextEdit.app/Contents/Resources/txt.icns"},
		})
	}

	got, err := gen.Query([]string{".*/bbb.*", ".*b.*"}, ":", []string{""})
	if err != nil {
		t.Errorf("unexpected error on query: %v\n", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}
