package input

import (
	"regexp"
	"strconv"
	"strings"
)

var splitter = regexp.MustCompile("([^@#:]*)([@#:]?)(.*)")

// chunkInput divides the given input into its before and after
// separate portions.
func chunkInput(s string) (string, string, string) {
	matches := splitter.FindAllStringSubmatch(s, -1)

	if matches == nil {
		return "", "", ""
	}
	return matches[0][1], matches[0][2], matches[0][3]
}

// fileExp transforms the given string into a regexp string
// appropriate for file patterns. Not expected to generate
// rational output on empty strings.
func fileExp(s string) string {
	subpaths := strings.Split(s, "/")
	fuzzedpaths := make([]string, 0, len(subpaths))
	for _, sp := range subpaths {
		ex := strings.Split(sp, "")
		fuzzedpaths = append(fuzzedpaths, strings.Join(ex, ".*"))
	}
	return ".*" + strings.Join(fuzzedpaths, "/") + ".*"
}

// fuzzyMatchers generates a sequence of matches of differing
// fuzziness and returns the set of regexp strings.
func fuzzyMatchers(s string) []string {
	m := make([]string, 0)

	// Filename only.
	m = append(m, s+"[^/]*$")

	// Anywhere in the path.
	m = append(m, s)

	// Complete sub-paths rooted.
	subpaths := strings.Split(s, "/")
	m = append(m, "^"+strings.Join(subpaths, "[^/]*/"))

	// Complete sub-paths, not rooted.
	m = append(m, strings.Join(subpaths, "[^/]*/[^/]*"))

	m = append(m, fileExp(s))
	return m
}

func inLineExp(s string) string {
	return ".*" + s + ".*"
}

// symbolExp returns a regexp to find symbols in Golang source.
func symbolExp(s string) string {
	ex := strings.Split(s, "")
	return "(func|type|var|const).*" + strings.Join(ex, "[a-zA-Z_0-9]*") + "[a-zA-Z_0-9]*"
}

// Parse generates query-language specific regexps and a query type.
func Parse(s string) ([]string, string, string) {
	prefix, sep, suffix := chunkInput(s)
	switch sep {
	case "@":
		return fuzzyMatchers(prefix), "/", symbolExp(suffix)
	case "#":
		return fuzzyMatchers(prefix), ":", numCheck(suffix)
	case ":":
		if suffix[0] == '/' {
			return fuzzyMatchers(prefix), "/", inLineExp(suffix[1:])
		}
		return fuzzyMatchers(prefix), ":", numCheck(suffix)
	case "":
		return fuzzyMatchers(prefix), ":", ""
	}
	return []string{""}, "", ""
}

func numCheck(s string) string {
	_, err := strconv.Atoi(s)
	if err != nil {
		return ""
	}
	return s
}
