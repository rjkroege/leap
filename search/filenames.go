package search

import (
	"bytes"
	"path/filepath"

	"github.com/rjkroege/leap/output"
)

// bytesliceify converts the given array of strings
// into an array of byte slices.
func bytesliceify(paths []string) [][]byte {
	t := make([][]byte, 0, len(paths))
	for _, s := range paths {
		if s[len(s)-1] != '/' {
			s = s + "/"
		}
		t = append(t, []byte(s))
	}
	return t
}

// Chops off the prefix. Fails if any one path is a prefix of another
// path. But that's silly. The prefix can come either from the
// configuration or from the base directories of the index.
func (ix *Search) trimmer(fs []byte) []byte {
	if ix.trimpaths == nil {
		// Cache the trimpaths to improve performance (per
		// profiling data.)
		if ix.prefixes != nil {
			ix.trimpaths = bytesliceify(ix.prefixes)
		} else {
			ix.trimpaths = bytesliceify(ix.Paths())
		}
		// TODO(rjk): most fs share the same path
		// prefix. We are doing way too much work.
	}
	paths := ix.trimpaths

	for _, p := range paths {
		// Probably wrong on windows.
		fs = bytes.TrimPrefix(fs, p)
	}
	return fs
}

func extend(base, suffix string) string {
	if suffix != "" {
		return base + ":" + suffix
	}
	return base
}

func (ix *Search) filenameResult(fnames []uint32, suffix string) ([]output.Entry, error) {
	// TODO(rjk): Consider a better way to find the pretty sub-name:
	// such as the shortest unique prefix.
	oo := make([]output.Entry, 0, MaximumMatches)

	for _, fn := range fnames {
		name := ix.NameBytes(fn)
		sname := string(name)
		title := filepath.Base(sname)

		oo = append(oo, output.Entry{
			Uid:   sname,
			Arg:   extend(sname, suffix),
			Title: extend(title, suffix),
			// Makes a temporary string.
			SubTitle: extend(string(ix.trimmer(name)), suffix),

			Type: "file:skipcheck",
			Icon: output.AlfredIcon{
				Filename: determineIconString(sname),
			},
		})
	}
	return oo, nil
}
