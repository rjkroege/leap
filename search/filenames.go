package search

import (
	"path/filepath"
	"strings"

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
			fs = strings.TrimPrefix(fs, p+"/")
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

func (ix *trigramSearch) filenameResult(fnames []uint32, suffix string) ([]output.Entry, error) {
	// TODO(rjk): Consider a better way to find the pretty sub-name:
	// such as the shortest unique prefix.
	oo := make([]output.Entry, 0, MaximumMatches)

	for _, fn := range fnames {
		name := ix.Name(fn)
		title := filepath.Base(name)

		oo = append(oo, output.Entry{
			Uid:      name,
			Arg:      extend(name, suffix),
			Title:    extend(title, suffix),
			SubTitle: extend(ix.trimmer(name), suffix),

			Type: "file",
			Icon: output.AlfredIcon{
				Filename: determineIconString(name),
			},
		})
	}
	return oo, nil
}
