package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
)

// TODO(rjk): Add the necessary flags to configure the project
// root and what not, force re-indexing etc.
var (
	ip = flag.Int("flagname", 1234, "help message for flagname")
)

func main() {
	flag.Usage = func() {
	        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0] )
	        fmt.Fprintf(os.Stderr, "	%s <flags listed below> <search string>\n", os.Args[0] )
	        flag.PrintDefaults()
	}

	flag.Parse()

	log.Printf("hi there %d, %v", *ip, flag.Arg(0))
	
	gen := search.NewStaticGenerator()
	entries, _ := gen.Query(flag.Arg(0))
	output.WriteOut(os.Stdout, entries)
}

