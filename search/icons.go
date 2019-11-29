package search

import (
	"path"
)

// determineIconString returns the name of a icon that we will
// use for this file type.
// Apple moves things around so make sure that the icons are in
// the leap workflow directory or in an app that is known to have been
// installed.
// Icon source is ~/Documents/icons-for-programming-languages.
// https://github.com/exercism/meta/issues?page=2&q=is%3Aissue+is%3Aclosed is a
// great reference for icons for various programming languages.
// The Alfred workflow is /Users/rjkroege/lib/Alfred.alfredpreferences/workflows/user.workflow.7C73B7F2-0E9A-40B1-94E7-9059936FBE13
func determineIconString(name string) string {
	// TODO(rjk): Make this case invariant?
	switch path.Ext(name) {
	case ".cpp", ".cc":
		return "cpp_logo.png"
	case ".h", ".hpp":
		return "h_logo.png"
	case ".css":
		return "/Applications/Safari.app/Contents/Resources/css.icns"
	case ".js":
		return "js.png"
	case ".java":
		return "java.png"
	case ".html", ".htm":
		return "/Applications/Safari.app/Contents/Resources/html.icns"
	case ".md", ".markdown":
		// This works iff I have Marked2 installed.
		return "/Applications/Marked 2.app/Contents/Resources/DocumentIcon.icns"
	case ".go":
		return "golang.icns"
	case ".py":
		return "python-logo-generic.png"
	case ".rust":
		return "rust.png"
	case ".swift":
		return "swift.png"
	case ".text", ".txt":
		return "/Applications/TextEdit.app/Contents/Resources/txt.icns"
	}
	// TODO(rjk): OWNERS, gn, DEPS, objective C

	return ""
}
