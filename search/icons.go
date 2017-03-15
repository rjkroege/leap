package search

import (
	"path"
)

// determineIconString returns the name of a icon that we will
// use for this file type.
func determineIconString(name string) string {

	switch path.Ext(name) {
	case ".cpp", "cc":
		return 	"/Applications/Xcode.app/Contents/Resources/c-plus-plus-source_Icon.icns"
	case ".h", ".hpp":
		return "/Applications/Xcode.app/Contents/Resources/c-header_Icon.icns"
	case ".css":
		return "/Applications/Safari.app/Contents/Resources/css.icns"
	case ".js":
		return "/Applications/Safari.app/Contents/Resources/js.icns"
	case ".html", ".htm":
		return "/Applications/Safari.app/Contents/Resources/html.icns"
	case ".md", ".markdown":
		return "/Applications/Marked 2.app/Contents/Resources/DocumentIcon.icns"
	case ".go":
		return "golang.icns"
	}
	return ""
}
