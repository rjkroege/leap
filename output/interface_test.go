package output

import (
	"bytes"
	"testing"
)

func TestWriteOutOnet(t *testing.T) {
	buffy := new(bytes.Buffer)

	testEntries := []Entry{
		{
			Arg:      "arg",
			Title:    "title",
			SubTitle: "sub-title",
			Icon: AlfredIcon{
				Filename: "filename",
				Type:     "fileicon",
			},
		},
	}

	expected := "<?xml version=\"1.0\"?>\n<items>\n\t<item arg=\"arg\">\n\t\t<title>title</title>\n\t\t<subtitle>sub-title</subtitle>\n\t\t<icon type=\"fileicon\">filename</icon>\n\t</item>\n</items>\n"

	if err := WriteOut(buffy, testEntries); err != nil {
		t.Errorf("unexpected error writing %v: %v", testEntries, err)
	}
	if got := buffy.String(); got != expected {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}
}

func TestWriteOutExtra(t *testing.T) {
	buffy := new(bytes.Buffer)
	testEntries := []Entry{
		{
			Arg:      "arg",
			Title:    "title",
			SubTitle: "sub-title",
			Icon: AlfredIcon{
				Filename: "filename",
				Type:     "fileicon",
			},
			AutoComplete: "auto-complete",
		},
	}

	expected := "<?xml version=\"1.0\"?>\n<items>\n\t<item arg=\"arg\" autocomplete=\"auto-complete\">\n\t\t<title>title</title>\n\t\t<subtitle>sub-title</subtitle>\n\t\t<icon type=\"fileicon\">filename</icon>\n\t</item>\n</items>\n"
	if err := WriteOut(buffy, testEntries); err != nil {
		t.Errorf("unexpected error writing %v: %v", testEntries, err)
	}
	if got := buffy.String(); got != expected {
		t.Errorf("got %#v exepcted %#v", got, expected)
	}

}
