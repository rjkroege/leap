package input

import (
	"regexp"
	"strconv"
	"strings"
)

var splitter = regexp.MustCompile("([^@/#:]*)([@/#:]?)([^@/#:]*)")

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
	ex := strings.Split(s, "")
	return ".*" + strings.Join(ex, ".*") + ".*"
}

// symbolExp returns a regexp to find symbols in Golang source.
func symbolExp(s string) string {
	ex := strings.Split(s, "")
	return "(func|type|var|const).*" + strings.Join(ex, "[a-zA-Z_0-9]*") + "[a-zA-Z_0-9]*"
}

// Parse generates query-language specific regexps and a query type.
func Parse(s string) (string, string, string) {
	prefix, sep, suffix := chunkInput(s)
	switch sep {
	case "@":
		return fileExp(prefix), "/", symbolExp(suffix)
	case "#", ":":
		return fileExp(prefix), ":", numCheck(suffix)
	case "/":
		return fileExp(prefix), "/", fileExp(suffix)
	case "":
		return fileExp(prefix), ":", ""
	}
	return "", "", ""
}

func numCheck(s string) string {
	_, err := strconv.Atoi(s)
	if err != nil {
		return ""
	}
	return s
}
