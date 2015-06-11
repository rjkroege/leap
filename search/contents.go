package search

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/google/codesearch/index"
	"github.com/google/codesearch/regexp"

	"github.com/rjkroege/leap/output"
)


const MaximumMatches = 50

type trigramSearch struct {
	index.Index
}

func NewTrigramSearch() output.Generator {
	return &trigramSearch{ *index.Open(index.File()) }
}

func (ix *trigramSearch) Query(fn, qtype, suffix string) ([]output.Entry, error) {
	// Don't know how to deal with other types other than line-no mode.
	log.Printf("Query: %#v", qtype)
	if qtype == ":" {
		return ix.fileQuery(fn, qtype, suffix)
	}

	// compile the in-file regexp
	pat := "(?m)" + suffix
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}

	//	compile the regexp
	log.Println("fn", fn)
	fre, err := regexp.Compile(fn)
	if err != nil {
		return nil, err
	}

	q := index.RegexpQuery(re.Syntax)
	post := ix.PostingQuery(q)
	
	fnames := make([]uint32, 0, len(post))

	for _, fileid := range post {
		name := ix.Name(fileid)
		sname := ix.trimmer(name)

		if fre.MatchString(sname, true, true) < 0 {
			continue
		}
		fnames = append(fnames, fileid)
	}


	log.Printf("finished the trigram query")

	matches := multiFile(fnames, re, ix)

	oo := make([]output.Entry,0, len(matches))

	// TODO(rjk): Do a better job of handling long strings.
	for _, m := range matches {
		name := m.fn
		// It would be nice if Alfred supported styled strings?
		title := fmt.Sprintf("%s:%d %s", filepath.Base(name), m.lineno, m.matchLine)
		
		oo = append(oo, output.Entry{
			Uid: name +  "/" + m.matchLine,
			Arg: fmt.Sprintf("%s:%d", name, m.lineno),
			Title: title,
			SubTitle:	fmt.Sprintf("%s:%d %s", ix.trimmer(name), m.lineno, m.matchLine),
			Type: "file",
			Icon: output.AlfredIcon{
				Filename: "blah.png",
			},
		})

		log.Printf("matched %#v", name)
	}

	return oo, nil
}