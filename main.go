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
	ip      = flag.Int("flagname", 1234, "help message for flagname")
	testlog = flag.Bool("testlog", false, "Log in the conventional way for running in a terminal")
	server = flag.Bool("server", false, "Run as a server. If a server is already running, does nothing.")
	stop = flag.Bool("stop", false, "Connect to the configured server and stop it.")
	host = flag.String("host", "", "Configure hostname for server. Empty host is short-circuited to operate in-memory.")
	indexpath = flag.String("indexpath", "",
		"Configure the path to the index file. Use CSEARCHINDEX if not provided.")
	
)

func LogToTemp() func() {
	logFile, err := ioutil.TempFile("/tmp", "leap")
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
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "	%s <flags listed below> <search string>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if !*testlog {
		defer LogToTemp()()
	}

	if *server {
		fmt.Fprintln(os.Stderr, "go run as server")
		return
	}
	if *stop {
		fmt.Fprintln(os.Stderr, "stop the running server")
		return
	}
	if *host != "" {
		fmt.Fprintln(os.Stderr, "set host field", *host)
		return
	}

	// Default mode of operation.
	// TODO(rjk): read configuration.


	fn, stype, suffix := input.Parse(flag.Arg(0))
	gen := search.NewTrigramSearch()

	// TODO(rjk): error check
	entries, _ := gen.Query(fn, stype, suffix)
	output.WriteOut(os.Stdout, entries)
}
