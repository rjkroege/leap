package search

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/codesearch/index"
	"github.com/google/codesearch/regexp"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
)

const MaximumMatches = 50

type trigramSearch struct {
	index.Index
	prefixes []string
	trimpaths [][]byte
}

func (ix *trigramSearch) filterFileIndicesForRegexpMatch(post []uint32, fre *regexp.Regexp, fnames []uint32, dedup map[uint32]struct{}) []uint32 {

	// re-process file names.
	for i := 0; len(fnames) < MaximumMatches && i < len(post); i++ {
		fileid := post[i]

		if _, ok := dedup[fileid]; ok {
			continue
		}

		// TODO(rjk): I am redoing (very cheap) work later.
		name := ix.NameBytes(fileid)
		sname := ix.trimmer(name)

		if fre.Match(sname, true, true) < 0 {
			continue
		}
		fnames = append(fnames, fileid)
		dedup[fileid] = struct{}{}
	}
	return fnames
}

// NewTrigramSearch returns a Generator that can search
// inside of files using index at path and project truncation
// prefixes.
func NewTrigramSearch(path string, prefixes []string) output.Generator {
	return &trigramSearch{*index.Open(path), prefixes, nil}
}

func (ix *trigramSearch) Query(fnl []string, qtype string, suffixl []string) ([]output.Entry, error) {
	suffix := suffixl[0]

	stime := time.Now()
	defer func(){
		log.Printf("query %v, %v, %v tool %v", fnl, qtype, suffixl, time.Since(stime))
	}()

	// TODO(rjk): code seems vaguely unclean
	// Produce a list of filename, all or content-matches only.
	var query *index.Query
	var re *regexp.Regexp

	// idea: I want some easy way to bound the number of responses
	// I can look at the search complexity and switch to regep mode if
	// it's insufficiently complicated.
	if qtype == ":" {
		query = &index.Query{Op: index.QAll}
	} else {
		pat := "(?m)" + suffix
		var err error
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

	fnames := make([]uint32, 0, MaximumMatches)
	dedup := make(map[uint32]struct{}, MaximumMatches)
	for _, fn := range fnl {
		//	compile the filename regexp
		fre, err := regexp.Compile(fn)
		if err != nil {
			return nil, err
		}
		fnames = ix.filterFileIndicesForRegexpMatch(post, fre, fnames, dedup)
	}

	if qtype == ":" {
		return ix.filenameResult(fnames, suffix)
	} else {
		return ix.contentSearchResult(fnames, re)
	}
}

// findLongestPrefix determines the length of the longest
// common prefix of the provided array of strings.
func findLongestPrefix(names []string) int {
	if len(names) < 1 {
		return 0
	}
	readers := make([]*strings.Reader, 0, len(names))
	for _, m := range names {
		readers = append(readers, strings.NewReader(m))
	}

	for {
		rune0, _, err := readers[0].ReadRune()
		if err != nil {
			return int(readers[0].Size())
		}
		for _, re := range readers[1:] {
			r, _, err := re.ReadRune()
			if err != nil {
				return int(re.Size())
			}
			if r != rune0 {
				re.UnreadRune()
				return int(re.Size()) - re.Len()
			}
		}
	}
	return 0
}

// nicelyTrimPath adjusts the given absolute path fn for
// informative visual display by removing unnecessary
// path components. fn is an absolute path, text before
// cut should be discarded.
func (ix *trigramSearch) nicelyTrimPath(fn []byte, cut int) string {
	// Adjust for leading / after cutting.
	if cut > 0 && fn[cut] == filepath.Separator && cut < len(fn) {
		cut++
	}
	cutstring := fn[cut:]
	trimstring := ix.trimmer(fn)
	if len(cutstring) < len(trimstring) {
		return ".../" + string(cutstring)
	}
	return ".../" + string(trimstring)
}

func (ix *trigramSearch) contentSearchResult(fnames []uint32, re *regexp.Regexp) ([]output.Entry, error) {
	// Search inside the files.
	matches := multiFile(fnames, re, ix)

	bn := make([]string, 0, len(matches))
	for _, m := range matches {
		bn = append(bn, filepath.Dir(m.fn))
	}
	trimpoint := findLongestPrefix(bn)

	oo := make([]output.Entry, 0, len(matches))

	for _, m := range matches {
		name := m.fn
		// It would be nice if Alfred supported styled strings. Then, I
		// could highlight the search results.
		title := fmt.Sprintf("%d %s", m.lineno, m.matchLine)
		arg := fmt.Sprintf("/%s:%d%s", base.Prefix, m.lineno, name)

		oo = append(oo, output.Entry{
			Uid:      fmt.Sprintf("%s:%d", name, m.lineno),
			Arg:      arg,
			Title:    title,
			SubTitle: fmt.Sprintf("%s:%d %s", ix.nicelyTrimPath([]byte(name), trimpoint), m.lineno, m.matchLine),
			Type:     "file",
			Icon: output.AlfredIcon{
				Filename: determineIconString(name),
			},
		})

		// Copy the content to the prefix.
		dir := filepath.Dir(arg)
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Println(dir, err)
			continue
		}
		if err := fileCopy(name, arg); err != nil {
			log.Println(name, arg, err)
			continue
		}
	}
	return oo, nil
}

func fileCopy(a, b string) error {
	sFile, err := os.Open(a)
	if err != nil {
		return err
	}
	defer sFile.Close()

	eFile, err := os.Create(b)
	if err != nil {
		return err
	}
	defer eFile.Close()

	_, err = io.Copy(eFile, sFile) // first var shows number of bytes
	if err != nil {
		return err
	}
	return nil
}
