package search

import (
	"log"
	"strings"

	"github.com/rjkroege/leap/output"
)

type staticGenerator []output.Entry

func NewStaticGenerator() output.Generator {
	return staticGenerator([]output.Entry{
		output.Entry{
			Uid:      "foo1",
			Arg:      "anArg",
			Type:     "type1",
			Title:    "Title",
			SubTitle: "the subtitle",
			Icon: output.AlfredIcon{
				Filename: "blah.png",
			},
		},
		output.Entry{
			Uid:      "foo2",
			Arg:      "anDifferentArg",
			Type:     "type2",
			Title:    "Another Title",
			SubTitle: "the other subtitle",
			Icon: output.AlfredIcon{
				Filename: "bling.png",
			},
		},
		output.Entry{
			Uid:      "foo3",
			Arg:      "/Users/rjkroege/tools/gopkg/src/github.com/rjkroege/leap/main.go",
			Type:     "type3",
			Title:    "Another Different Title",
			SubTitle: "a source file",
			Icon: output.AlfredIcon{
				Filename: "bling.png",
			},
			Valid: "YES",
		},
	})
}

// Presumption is that s is a query. So is probably a regexp.
func (sg staticGenerator) Query(fn, qtype, suffix string) ([]output.Entry, error) {
	s := fn

	// basic approach: prefix match
	result := make([]output.Entry, 0)

	for _, e := range sg {
		if strings.HasPrefix(e.Title, s) {
			result = append(result, e)
		}
	}

	return result, nil
}
