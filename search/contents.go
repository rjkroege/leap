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
	prefixes []string
}

func (ix *trigramSearch) filterFileIndicesForRegexpMatch(post []uint32, fre*regexp.Regexp) []uint32 {
	fnames := make([]uint32, 0, len(post))

	// re-process file names.
	for _, fileid := range post {
		name := ix.Name(fileid)
		sname := ix.trimmer(name)

		if fre.MatchString(sname, true, true) < 0 {
			continue
		}
		fnames = append(fnames, fileid)
	}
	return fnames
}

// NewTrigramSearch returns a Generator that can search
// inside of files using index at path and project truncation
// prefixes.
func NewTrigramSearch(path string, prefixes []string) output.Generator {
	return &trigramSearch{*index.Open(path), prefixes}
}

func (ix *trigramSearch) Query(fn, qtype, suffix string) ([]output.Entry, error) {
	log.Printf("Query: %#v", qtype)

	//	compile the filename regexp
	log.Println("fn", fn)
	fre, err := regexp.Compile(fn)
	if err != nil {
		return nil, err
	}

	// TODO(rjk): code seems vaguely unclean
	// Produce a list of filename, all or content-matches only.
	var query *index.Query
	var re *regexp.Regexp

	if qtype == ":" {
		query = &index.Query{Op: index.QAll}
	} else {
		pat := "(?m)" + suffix
		re, err = regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		query = index.RegexpQuery(re.Syntax)
		// TODO(rjk): The result of this is that we first build a list of
		// filenames. Bound in some way.

		// TODO(rjk): if the number of files specified by the file regexp
		// lies below some threshold, then skip the full index search.
	}
	post := ix.PostingQuery(query)

	// TODO(rjk): abundant fuzzy matching improvements here.
	fnames := ix.filterFileIndicesForRegexpMatch(post, fre)

	if qtype == ":" {
		return ix.filenameResult(fnames, suffix)
	} else {
		return ix.contentSearchResult(fnames, re)
	}
}

func (ix *trigramSearch) contentSearchResult(fnames []uint32, re *regexp.Regexp) ([]output.Entry, error) {
	// Search inside the files.
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
