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

// NewTrigramSearch returns a Generator that can search
// inside of files.
func NewTrigramSearch() output.Generator {
	return &trigramSearch{*index.Open(index.File())}
}

func (ix *trigramSearch) Query(fn, qtype, suffix string) ([]output.Entry, error) {
	// TODO(rjk): Refactor this code.
	if qtype == ":" {
		return ix.fileQuery(fn, qtype, suffix)
	}
	log.Printf("Query: %#v", qtype)

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

	matches := multiFile(fnames, re, ix)

	oo := make([]output.Entry, 0, len(matches))

	// TODO(rjk): Do a better job of displaying long match strings.
	// In particular, preserve the function name in function matches.
	for _, m := range matches {
		name := m.fn
		// It would be nice if Alfred supported styled strings. Then, I
		// could highlight the search results.
		title := fmt.Sprintf("%s:%d %s", filepath.Base(name), m.lineno, m.matchLine)

		oo = append(oo, output.Entry{
			Uid:      name + "/" + m.matchLine,
			Arg:      fmt.Sprintf("%s:%d", name, m.lineno),
			Title:    title,
			SubTitle: fmt.Sprintf("%s:%d %s", ix.trimmer(name), m.lineno, m.matchLine),
			Type:     "file",
			Icon: output.AlfredIcon{
				Filename: name,
				Type: "fileicon",
			},
		})
	}

	return oo, nil
}
