package output

// Generates a list of candidate entries based on a combination of
// context and a provided argument

import (
	"encoding/xml"
	"io"
)

type AlfredIcon struct {
	Filename string `xml:",chardata"`
	Type     string `xml:"type,attr,omitempty"`
}

type Entry struct {
	XMLName      xml.Name   `xml:"item"`
	Uid          string     `xml:"uid,attr,omitempty"`
	Arg          string     `xml:"arg,attr"`
	Type         string     `xml:"type,attr,omitempty"`
	Valid        string     `xml:"valid,attr,omitempty"`
	AutoComplete string     `xml:"autocomplete,attr,omitempty"`
	Title        string     `xml:"title"`
	SubTitle     string     `xml:"subtitle"`
	Icon         AlfredIcon `xml:"icon"`
}

type Generator interface {
	// Query searches for the specified fn (file name) patterns and
	// suffix (in content) patterns. The patterns should be arranged
	// in descending order of desirability. Entires satisfying the set
	// of search queries are returned or error.
	Query(fn []string, qtype string, suffix []string) ([]Entry, error)
}

type items struct {
	Items []Entry
}

func WriteOut(w io.Writer, e []Entry) error {
	if _, err := io.WriteString(w, "<?xml version=\"1.0\"?>\n"); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "	")
	if err := enc.Encode(items{e}); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}
