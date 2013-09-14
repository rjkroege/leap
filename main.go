package main

import (
	"bufio"
	"code.google.com/p/goplan9/plan9"
	"code.google.com/p/goplan9/plan9/client"
	"fmt"
	"github.com/raguay/goAlfred"
	"log"
	"os"
	"regexp"
	"strconv"
	"unicode/utf8"
)


// Keek pristine with goal of upstreaming.
type WinEntry struct {
	Id uint
	Taglen uint
	Bodylen uint
	Isdir bool
	Ismod bool
	Filename string	
	Tag string	
}

const maxCharactersInName = 65

/*
	Tacky elision. Might want to do something nicer. Like strip the common prefixes.
*/
func ElidedFileName(n string) string {
	if len(n) > maxCharactersInName {
		return "..." + n[len(n) - maxCharactersInName:]
	}
	return n
}

// rip icons out of the peepcode app?
//func PickIcon(n string) string {
//}

func MakeRegexp(s string, prefix string, suffix string) (*regexp.Regexp, error) {
	b := make([]byte, 0, len(s) * 4);
	for i, r := range(s) {
		b = append(b, prefix...)
		b = append(b, s[i:i+utf8.RuneLen(r)]...)
		b = append(b, suffix...)
		// b = append(b, "[^/]*"...)
	}
	log.Println(string(b))
	return regexp.Compile(string(b))
}


/*
	Adds an entry to the Alfred results.
*/
func AddResultEntry(i int, w *WinEntry) {
	
	goAlfred.AddResult(strconv.FormatInt(int64(i), 10), w.Filename, ElidedFileName(w.Filename), w.Filename, "icon.png", "yes", "", "")
}


func main() {

	// TODO(rjkroege): note that we need to do this based on the argument... 
	// eventually we'll have a buffer specified...
	// but I should probably always check that the buffer is in the index
	// TODO(rjkroege): worry about the error handling.
	fsys, err := client.MountService("acme")
	if err != nil {
		log.Fatal("can't attach to the acme: " + err.Error())
	}

	fid, err := fsys.Open("index", plan9.OREAD)
	if err != nil {
		log.Fatal("can't open the index file: " + err.Error())
	}

	token_count := 0	
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		switch {
		case token_count < 6:
			advance, token, err = bufio.ScanWords(data, atEOF)
			token_count++
		case token_count == 6:
			advance, token, err = bufio.ScanLines(data, atEOF)
			token_count = 0;	
		}		
            return				
	}

	wins := make([]*WinEntry,0, 10)		
	scanner := bufio.NewScanner(fid)
	scanner.Split(split)

	for a := make([]string, 0, 7); scanner.Scan(); {
		a = append(a, scanner.Text())
		if len(a) == 7 {
			ip := make([]uint, 5, 5)
			for i:= 0; i < 5; i++ {
				p, _ := strconv.ParseUint(a[i], 10, 32)
				ip[i] = uint(p)
			}
			wins = append(wins, &WinEntry{ip[0], ip[1], ip[2], ip[3] == 1, ip[4] == 1, a[5], a[6]})
			a = make([]string, 0, 7)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	// Output...
//	for _, w := range(wins) {
//		log.Print(w)
//	}

	/*
		How does this all work anyway. 
		Search order...

		a search always begins with a file name. or : if to use the current file. We don't know the current file. It
		would be nice to figure out the current file. I don't think that acme will tell me where the cursor is...

	*/



	// What about spaces? I can use this for something? Why don't I just smush the arguments together?
	if len(os.Args) > 1 {
		// Could loop over the regexps.
		rep, err  := MakeRegexp(os.Args[1], "", "[^/]*")
		if err != nil {
			log.Print("MakeRegexp, exact: ", err.Error())
		}
		ref, err  := MakeRegexp(os.Args[1], "", ".*")
		if err != nil {
			log.Print("MakeRegexp, exact: ", err.Error())
		}

		for i, w := range(wins) {
			log.Println("regexp testing: " + w.Filename)
			if rep.MatchString(w.Filename) {
				log.Println("regexp matched: " + w.Filename)
				AddResultEntry(i, w)
			}
		}

		// no need to list fid again right?

		for i, w := range(wins) {
			log.Println("regexp testing: " + w.Filename)
			if ref.MatchString(w.Filename) {
				log.Println("regexp matched: " + w.Filename)
				AddResultEntry(i, w)
			}
		}
	} 


/*
	// original example code
	if(len(os.Args) > 1) {
		switch (os.Args[1]) {
			case "1": 
				goAlfred.AddResult("testUID1", "test argument1", "This is my title1", "test substring1", "icon.png", "yes", "", "")
				goAlfred.AddResult("testUID2", "test argument2", "This is my title2", "test substring2", "icon.png", "yes", "", "")
				goAlfred.AddResult("testUID3", "test argument3", "This is my title3", "test substring3", "icon.png", "yes", "", "")
			case "2":	
				goAlfred.AddResult("testUID2", "test argument2", "This is my title2", "test substring2", "icon.png", "yes", "", "")
				goAlfred.AddResult("testUID1", "test argument1", "This is my title1", "test substring1", "icon.png", "yes", "", "")
				goAlfred.AddResult("testUID3", "test argument3", "This is my title3", "test substring3", "icon.png", "yes", "", "")
			case "3":
				goAlfred.AddResult("testUID3", "test argument3", "This is my title3", "test substring3", "icon.png", "yes", "", "")
				goAlfred.AddResult("testUID1", "test argument1", "This is my title1", "test substring1", "icon.png", "yes", "", "")
				goAlfred.AddResult("testUID2", "test argument2", "This is my title2", "test substring2", "icon.png", "yes", "", "")

		}
	} else {
		goAlfred.AddResult("testUID3", "test argument3", "This is my title3", "test substring3", "icon.png", "yes", "", "")
		goAlfred.AddResult("testUID", "test argument", "This is my title", "test substring", "icon.png", "yes", "", "")
		goAlfred.AddResult("testUID2", "test argument2", "This is my title2", "test substring2", "icon.png", "yes", "", "")		
	}
*/

	//
	// Print out the created XML. 
	//
	fmt.Print(goAlfred.ToXML())
}