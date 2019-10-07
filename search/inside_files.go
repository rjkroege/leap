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
		buf = buf[:len(buf)+n]
		end := len(buf)
		if err == nil {
			end = bytes.LastIndex(buf, nl) + 1
		} else {
			endText = true
		}

		chunkStart := 0
		for chunkStart < end {
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
	orderingchans := make([]chan int, len(fnames))
	resultchan := make(chan []*inFileMatches)

	for i := range orderingchans {
		ochan := make(chan int)
		orderingchans[i] = ochan

		go func(c chan int, name, pat string) {
			// regexp.Regexp is not threadsafe so make one.
			// We know it will compile because we already compiled it.
			re, _ = regexp.Compile(pat)

			// Do concurrent work. Some might be wasted.
			m, err := searchInFile(re, name)
			if err != nil {
				log.Println("multiFile error: ", err)
				// Be sure to ship back an empty array.
				m = []*inFileMatches{}
			}

			// Block here until either invoker says yes or no.
			if _, ok := <-c; ok {
				resultchan <- m
			}
		}(ochan, ix.Name(fnames[i]), re.String())
	}

	for i := 0; len(matches) < MaximumMatches && i < len(orderingchans); i++ {
		orderingchans[i] <- 1
		matches = append(matches, (<-resultchan)...)
	}

	for _, c := range orderingchans {
		close(c)
	}

	return matches
}
