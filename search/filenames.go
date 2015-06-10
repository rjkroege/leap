package search

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/google/codesearch/index"
	"github.com/google/codesearch/regexp"

	"github.com/rjkroege/leap/output"
)

// Chops off the prefix. Fails if any one path is a prefix
// of another path. But that's silly.
func (ix *trigramSearch) trimmer(fs string) string {

	for _, p := range ix.Paths() {
		fs = strings.TrimPrefix(fs, p)
	}
	return fs
}

func extend(base, suffix string) string {
	if suffix != "" {
		return base + ":" + suffix
	}
	return base
}


func (ix *trigramSearch) fileQuery(fn, qtype, suffix string) ([]output.Entry, error) {
	// Don't know how to deal with other types other than line-no mode.
	log.Printf("Query: %#v", qtype)
	if qtype != ":" {
		return nil, nil
	}

	//	compile the regexp
	log.Println("fn", fn)
	fre, err := regexp.Compile(fn)
	if err != nil {
		return nil, err
	}

	// allQuery inspired by example in Russ's code.
	allQuery := &index.Query{Op: index.QAll}
	post := ix.PostingQuery(allQuery)
	
	fnames := make([]uint32, 0, len(post))

	for _, fileid := range post {
		name := ix.Name(fileid)
		sname := ix.trimmer(name)

		if fre.MatchString(sname, true, true) < 0 {
			continue
		}
		fnames = append(fnames, fileid)
	}

	// Better way to find the pretty sub-name: shortest unique prefix

	oo := make([]output.Entry,0, 20)

	for i := 0; i < 20 && i < len(fnames); i++ {
		name := ix.Name(fnames[i])
		title := filepath.Base(name)
		
		oo = append(oo, output.Entry{
			Uid: name,
			Arg: extend(name, suffix),
			Title: extend(title, suffix),
			SubTitle:	extend(ix.trimmer(name), suffix),

			Type: "file",
			Icon: output.AlfredIcon{
				Filename: "blah.png",
			},
		})

		log.Printf("matched %#v", name)
	}

	return oo, nil
}