package search

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/codesearch/regexp"
)

type inFileMatches struct {
	fn        string
	lineno    int
	matchLine string
}

var nl = []byte{'\n'}

func countNL(b []byte) int {
	n := 0
	for {
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			break
		}
		n++
		b = b[i+1:]
	}
	return n
}

func searchInFile(re *regexp.Regexp, name string) ([]*inFileMatches, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	matches := make([]*inFileMatches, 0, MaximumMatches)

	var (
		buf       = make([]byte, 0, 1<<20)
		lineno    = 1
		beginText = true
		endText   = false
	)

	for {
		n, err := io.ReadFull(f, buf[len(buf):cap(buf)])
		log.Println("read chunk", len(buf), cap(buf), err)
		buf = buf[:len(buf)+n]
		end := len(buf)
		if err == nil {
			end = bytes.LastIndex(buf, nl) + 1
		} else {
			endText = true
		}

		chunkStart := 0
		for chunkStart < end {
			log.Println("inner loop")
			m1 := re.Match(buf[chunkStart:end], beginText, endText) + chunkStart
			beginText = false
			if m1 < chunkStart {
				break
			}
			lineStart := bytes.LastIndex(buf[chunkStart:m1], nl) + 1 + chunkStart
			lineEnd := m1 + 1
			if lineEnd > end {
				lineEnd = end
			}
			lineno += countNL(buf[chunkStart:lineStart])
			line := buf[lineStart:lineEnd]

			log.Println("len, cap ", len(matches), cap(matches))
			if len(matches) == cap(matches) {
				return matches, nil
			}

			matches = append(matches, &inFileMatches{
				fn:        name,
				lineno:    lineno,
				matchLine: string(line),
			})

			lineno++
			chunkStart = lineEnd
		}
		if err == nil {
			lineno += countNL(buf[chunkStart:end])
		}
		n = copy(buf, buf[end:])
		buf = buf[:n]
		if len(buf) == 0 && err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return matches, fmt.Errorf("%s: %v\n", name, err)
			}
			break
		}
	}
	return matches, nil
}

func multiFile(fnames []uint32, re *regexp.Regexp, ix *trigramSearch) []*inFileMatches {
	matches := make([]*inFileMatches, 0, MaximumMatches)
	for i := 0; len(matches) < cap(matches) && i < len(fnames); i++ {
		m, err := searchInFile(re, ix.Name(fnames[i]))
		if err != nil {
			log.Println("multiFile error: ", err)
		} else {
			matches = append(matches, m...)
		}
	}
	return matches
}
