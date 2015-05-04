package main

import (
	"os"
	"log"

	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
)


func main() {
	log.Printf("hi there")

	gen := search.NewStaticGenerator()

	// Need to read from the command line.
	entries, _ := gen.Query("hi")

	output.WriteOut(os.Stdout, entries)
}

