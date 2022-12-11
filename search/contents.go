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

type Search struct {
	name string
	index.Index
	prefixes  []string
	trimpaths [][]byte
}

func (ix *Search) GetName() string {
	return ix.name
}

// filterFileIndicesForRegexpMatch looks up each file index in the
// backing cindex store and adds it to the result list if its name
// matches the filename regexp pattern.
func (ix *Search) filterFileIndicesForRegexpMatch(post []uint32, re *regexp.Regexp, fnames []uint32) []uint32 {
	// This loop could conceivably be over all of the filenames. This could
	// be large. Keeping the body efficient has large impact.
	for i := 0; len(fnames) < MaximumMatches && i < len(post); i++ {
		fileid := post[i]

		name := ix.NameBytes(fileid)
		sname := ix.trimmer(name)

		if re.Match(sname, true, true) >= 0 {
			fnames = append(fnames, fileid)
			continue
		}
	}
	return fnames
}

// reorderMatchByFuzziness reorders the matches to be in increasing order
// of fuzziness so that best matches appear first.
func (ix *Search) reorderMatchByFuzziness(matches []uint32, fnls []string) ([]uint32, error) {
	res := make([]*regexp.Regexp, len(fnls))
	reordered := make([][]uint32, len(res))
	for i := range res {
		reordered[i] = make([]uint32, 0, len(matches))
		fre, err := regexp.Compile(fnls[i])
		if err != nil {
			return nil, err
		}
		res[i] = fre
	}

outer:
	for _, fileid := range matches {
		for i, re := range res[0 : len(res)-1] {
			name := ix.NameBytes(fileid)
			sname := ix.trimmer(name)
			if re.Match(sname, true, true) >= 0 {
				reordered[i] = append(reordered[i], fileid)
				continue outer
			}
		}
		reordered[len(res)-1] = append(reordered[len(res)-1], fileid)
	}

	result := reordered[0]
	for _, r := range reordered[1:] {
		result = append(result, r...)
	}
	return result, nil
}

// NewTrigramSearch returns a Generator that can search
// inside of files using index at path and project truncation
// prefixes.
func NewTrigramSearch(path string, prefixes []string) *Search {
	return &Search{path, *index.Open(path), prefixes, nil}
}

type ContentSearcher interface {
	ContentSearchResult(fnames []uint32, re *regexp.Regexp, suffix string) ([]output.Entry, error)
}

// Query searches for the specified fn (file name) patterns and suffix
// (in content) patterns. The patterns should be arranged in descending
// order of desirability. Entires satisfying the set of search queries
// are returned or error. The ContentSearcher implementation will be used
// to remote searches into the interior of files.
// TODO(rjk): divide this code into two functions based on suffix?
func (ix *Search) Query(fnl []string, qtype string, suffixl []string, cs ContentSearcher) ([]output.Entry, error) {
	suffix := suffixl[0]

	stime := time.Now()

	// TODO(rjk): code seems vaguely unclean
	// Produce a list of filename, all or content-matches only.
	var query *index.Query
	var re *regexp.Regexp
	pat := ""

	// TODO(rjk): Explore having different (extended) search properties.
	// In essence, I want to do the search as if the argument is a filename
	// or something in a file.
	// TODO(rjk): I want some easy way to bound the number of responses
	// I can look at the search complexity and switch to regexp mode if
	// it's insufficiently complicated.
	if qtype == ":" {
		// This is a filename-only search.
		query = &index.Query{Op: index.QAll}

		// The perf problem here is we don't have an index of the
		// filenames. But on a modern laptop, it's 12ms for Chrome.
		// That's fast enough for doing it for each typed character.
		// Chrome is an atypically large use case.
	} else {
		// This is a contents search. Warmish-runs take 17ms on Chrome.
		pat = "(?m)" + suffix
		var err error
		re, err = regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		query = index.RegexpQuery(re.Syntax)
	}
	post := ix.PostingQuery(query)

	// File tokens are 32 bit integers.
	fnames := make([]uint32, 0, MaximumMatches)

	// We merge the regexps together for fastest initial filter.
	melded := strings.Join(fnl, "|")
	fre, err := regexp.Compile(melded)
	if err != nil {
		return nil, err
	}

	// This is O(n) over the list of candidate files. That would be all of the
	// files for a file-name only match.
	fnames = ix.filterFileIndicesForRegexpMatch(post, fre, fnames)

	// Reorder the results for better quality.
	if fnames, err = ix.reorderMatchByFuzziness(fnames, fnl); err != nil {
		return nil, err
	}

	defer func() {
		log.Printf("Query %v, %v, %v tool outputstate total %v", fnl, qtype, suffixl, time.Since(stime))
	}()

	if qtype == ":" {
		// Filename results do not actually require the files.
		// If we have the index locally, we would appear to not
		// need to ask the remote for anything.
		return ix.filenameResult(fnames, suffix)
	} else {
		// Conversely, file search requires access to the files.
		// So if the files aren't actually local, we need to send
		// messages here. This is expensive for content searches
		// because it looks in each one.
		return cs.ContentSearchResult(fnames, re, pat)
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
func (ix *Search) nicelyTrimPath(fn []byte, cut int) string {
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

// contentSearchResult actually searches inside the files to confirm the
// index matches.
func (ix *Search) ContentSearchResult(fnames []uint32, re *regexp.Regexp, _ string) ([]output.Entry, error) {
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

		// TODO(rjk): make icons for C++ etc. work correctly here.
		// The golang icons work? They do. They're part of the workflow.
		// I can make custom icons, put in the workflow and save the copies.
		// This will simplify the "remote-i-fying"
		// Copy the content to the prefix so that icons work properly.
		// TODO(rjk): this doesn't work right for remote?
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

// TODO(rjk): Remove this code?
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
