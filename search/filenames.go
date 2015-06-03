package search

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/google/codesearch/index"
	"github.com/google/codesearch/regexp"

	"github.com/rjkroege/leap/output"
)

type filenameSearch struct {
	index.Index
}

func NewFileNameSearch() output.Generator {
	return &filenameSearch{ *index.Open(index.File()) }
}

// Chops off the 
func (ix *filenameSearch) trimmer(fs string) string {

	for _, p := range ix.Paths() {
		fs = strings.TrimPrefix(fs, p)
	}
	return fs
}

func (ix *filenameSearch) Query(fn, qtype, suffix string) ([]output.Entry, error) {
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
			Arg: name,
			Title: title,
			SubTitle:	ix.trimmer(name),

			Type: "file",
			Icon: output.AlfredIcon{
				Filename: "blah.png",
			},
		})

		log.Printf("matched %#v", name)
	}

	return oo, nil
}