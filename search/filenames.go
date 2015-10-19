package search

import (
	"path/filepath"
	"strings"

	"github.com/google/codesearch/index"
	"github.com/google/codesearch/regexp"

	"github.com/rjkroege/leap/output"
)

// Chops off the prefix. Fails if any one path is a prefix of another
// path. But that's silly. The prefix can come either from the
// configuration or from the base directories of the index.
func (ix *trigramSearch) trimmer(fs string) string {
	paths := ix.Paths()
	if ix.prefixes != nil {
		paths = ix.prefixes
	}
	
	for _, p := range paths {
		// Probably wrong on windows.
		if p[len(p)-1] == '/' {
			fs = strings.TrimPrefix(fs, p)
		} else {
			fs = strings.TrimPrefix(fs, p + "/")
		}
	}
	return fs
}

func extend(base, suffix string) string {
	if suffix != "" {
		return base + ":" + suffix
	}
	return base
}

func (ix *trigramSearch) fileQuery(fre *regexp.Regexp, fn, suffix string) ([]output.Entry, error) {
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

	// TODO(rjk): Consider a better way to find the pretty sub-name:
	// such as the shortest unique prefix.

	oo := make([]output.Entry, 0, 20)

	for i := 0; i < 20 && i < len(fnames); i++ {
		name := ix.Name(fnames[i])
		title := filepath.Base(name)

		oo = append(oo, output.Entry{
			Uid:      name,
			Arg:      extend(name, suffix),
			Title:    extend(title, suffix),
			SubTitle: extend(ix.trimmer(name), suffix),

			Type: "file",
			Icon: output.AlfredIcon{
				Filename: name,
				Type: "fileicon",
			},
		})
	}
	return oo, nil
}
