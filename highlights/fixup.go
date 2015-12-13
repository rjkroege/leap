// Package highlights provides a capability to modify the output of
// highlight so that desired line is highlighted for easier display.
package highlights

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
)

var lineRecognizer = regexp.MustCompile("^(<li )(.*)$")

const suffix = `<script>
document.getElementById("theline").scrollIntoViewIfNeeded(true);
</script>
`

func ShowDesiredLineInFile(lineno string, in io.Reader, out io.Writer) error {
	// log.Println("ShowDesiredLineInFile ", lineno)
	ln := -1
	if lineno != "" {
		ln, _ = strconv.Atoi(lineno)
	}

	wr := bufio.NewWriter(out)
	scanner := bufio.NewScanner(in)
	i := 1
	for scanner.Scan() {
		s := scanner.Text()
		matches := lineRecognizer.FindAllStringSubmatch(s, -1)
		// log.Printf("%d %#v\n", i, matches)

		switch {
		case matches == nil:
			wr.WriteString(s)
			wr.WriteRune('\n')
		case i != ln:
			i++
			wr.WriteString(s)
			wr.WriteRune('\n')
		case i == ln:
			i++
			wr.WriteString(matches[0][1])
			wr.WriteString(` id="theline" style="background-color: rgb(80,80,80);" `)
			wr.WriteString(matches[0][2])
			wr.WriteRune('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		wr.Flush()
		return err
	}
	if ln >= 1 {
		wr.WriteString(suffix)
	}
	wr.Flush()
	return nil
}
