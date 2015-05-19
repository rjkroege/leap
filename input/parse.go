package input

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

var splitter = regexp.MustCompile("([^@/#:]*)([@/#:]?)([^@/#:]*)")

// chunkInput divides the given input into its before and after
// separate portions.
func chunkInput(s string) (string, string, string) {
	// split the string

	matches := splitter.FindAllStringSubmatch(s, -1)
	log.Printf("matches: %v", matches)

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

// Go only. 
func symbolExp(s string) string {
	ex := strings.Split(s, "")
	return "(func|type|var|const).*" + strings.Join(ex, "[a-zA-Z_0-9]*") + "[a-zA-Z_0-9]*"
}


// func Parse(s string) (string, string, string) {
//}


func numCheck(s string) string {
	_, err := strconv.Atoi(s)
	if err != nil {
		return ""
	}
	return s
}




