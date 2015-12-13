// Package input provides primitives for processing input to leap
//
// This input can occur in two forms: input to the leap command
// when running in search mode or encoded path strings that need
// to be unencoded.
//
// Leap's search syntax: <path>[:number] | [<path>][:/<search>]. The
// path will filter the list of files to those that match (fuzzily) the
// provided path. The search string can be any valid regexp.
//
// Leap's output is intended to be used in an Alfred app workflow.
// Amongst other content, the arg value for each entry ends being
// used in two different fashions: as an argument to a follow-on
// workflow element that opens files and as a file path argument to
// the quicklook facility.
//
// The intended usage of leap is to provide arg strings that are valid
// input to the Plan9 plumber: <path>[:number]. If the number is
// present, these are not valid input to the quicklook as it uses the
// terminating extension to select which quicklook should be invoked
// and there is no quicklook registered to open files with a number
// suffix.
//
// Consequently, leap ships encoded paths: the number is embedded
// as a prefix at the beginning of the path. Note that this may overlap
// with a valid path string in / so is hard-coded to something unlikely
// to be found in the root directory of a UNIX system.
package input

import (
	"regexp"

	"github.com/rjkroege/leap/base"
)

var decoder = regexp.MustCompile("(.*" + base.Prefix + ":([0-9]+))?(.*)")

// EncodedToPlumb takes the given string and converts
// to a plumb string.
func EncodedToPlumb(s string) string {
	matches := decoder.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return ""
	}
	if matches[0][2] == "" {
		return matches[0][3]
	} else {
		return matches[0][3] + ":" + matches[0][2]
	}
}

// EncodedToFile takes the given string and removes
// the prefix, returning only the valid file path portion.
func EncodedToFile(s string) string {
	matches := decoder.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return ""
	}
	return matches[0][3]
}

// EncodedToNumber takes the given string and removes
// the prefix, returning only the desired line number.
func EncodedToNumber(s string) string {
	matches := decoder.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return ""
	}
	return matches[0][2]
}
