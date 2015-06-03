package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/rjkroege/leap/input"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
)

// TODO(rjk): Add the necessary flags to configure the project
// root and what not, force re-indexing etc.
var (
	ip = flag.Int("flagname", 1234, "help message for flagname")
	testlog = flag.Bool("testlog", false, "Log in the conventional way for running in a terminal")
)

func LogToTemp() func()() {

	logFile, err  := ioutil.TempFile("/tmp", "leap")
	if err != nil {
		log.Panic("leap couldn't make a logging file: %v", err)
	}

	log.SetOutput(logFile)

	return func() {
		log.SetOutput(os.Stderr)
	}
}

func main() {
	flag.Usage = func() {
	        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0] )
	        fmt.Fprintf(os.Stderr, "	%s <flags listed below> <search string>\n", os.Args[0] )
	        flag.PrintDefaults()
	}
	flag.Parse()

	if !*testlog {
		defer LogToTemp()()
	}

	log.Printf("hi there %d, %v", *ip, flag.Arg(0))

	fn, stype, suffix := input.Parse(flag.Arg(0))

	log.Println("parse out", fn, stype)

	
	// gen := search.NewStaticGenerator()
	gen := search.NewFileNameSearch()

	// TODO(rjk): error check
	entries, _ := gen.Query(fn, stype, suffix)
	output.WriteOut(os.Stdout, entries)
}

