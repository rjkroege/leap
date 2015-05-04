package search

import (
	"log"
	"strings"

	"github.com/rjkroege/leap/output"
)

type staticGenerator []output.Entry



func NewStaticGenerator() output.Generator {
	return staticGenerator( []output.Entry{
		output.Entry{
			Uid: "foo1",
			Arg: "anArg",
			Type: "type1"	,
			Title: "Title",
			SubTitle: "the subtitle",
			Icon: output.AlfredIcon{
				Filename: "blah.png",
			},
		},
		output.Entry{
			Uid: "foo2",
			Arg: "anDifferentArg",
			Type: "type2"	,
			Title: "Another Title",
			SubTitle: "the other subtitle",
			Icon: output.AlfredIcon{
				Filename: "bling.png",
			},
		},
	})
}


func (sg staticGenerator) Query(s string) ([]output.Entry, error) {
	log.Printf("query: %s", s)

	// basic approach: prefix match
	result := make([]output.Entry, 0)
	
	for _, e := range(sg) {
		if strings.HasPrefix(e.Title, s) {
			result = append(result, e)
		}
	}

	return result, nil	
}