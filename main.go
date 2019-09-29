package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/client"
	"github.com/rjkroege/leap/input"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
	"github.com/rjkroege/leap/server"

	// Uncomment to turn on profiling.
	// "github.com/pkg/profile"
)

// TODO(rjk): It is conceivable that I will want to support having a re-writing
// path configuration: client can re-write paths for where the file is to be found
// in the client. I currently don't have this issue. So I'm not going to deal.
var (
	testlog = flag.Bool("testlog", false,
		"Log in the conventional way for running in a terminal. Also changes where to find the configuration file.")
	runServer = flag.Bool("server", false, "Run as a server. If a server is already running, does nothing.")
	stop      = flag.Bool("stop", false, "Connect to the configured server and stop it.")

	index     = flag.Bool("index", false, "Connect to the configured server and ask it to re-index the configured path.")
	decodePlumb = flag.Bool("dp", false,
		"Decode the single provided path and convert it back into a valid plumb address")
)

func main() {
	// Uncomment to turn on profiling.
	// defer profile.Start().Stop()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "	%s <flags listed below> <search string>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if !*testlog {
		defer base.LogToTemp()()
	}

	switch {
	case *decodePlumb:
		log.Println("running decodePlumb")
		os.RemoveAll(base.SubPrefix)
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(0)
		}
		log.Println("output", input.EncodedToPlumb(flag.Arg(0)))
		fmt.Println(input.EncodedToPlumb(flag.Arg(0)))
		os.Exit(0)
	case *runServer:
		fmt.Fprintln(os.Stderr, "go run as server")
		config, err := base.GetConfiguration(base.Filepath(*testlog))
		if err != nil {
			log.Fatal("couldn't read configuration: ", err)
		}
		server.BeginServing(config)
		os.Exit(0)
	case *stop:
		config, err := base.GetConfiguration(base.Filepath(*testlog))
		if err != nil {
			log.Fatal("couldn't read configuration: ", err)
			return
		}
		if err := client.Shutdown(config); err != nil {
			log.Println("shutdown generated output: ", err)
		}
		os.Exit(0)
	}

	// May exit.
	base.UpdateConfigIfNecessary(flag.Args(), *testlog)
	os.RemoveAll(base.SubPrefix)

	config, err := base.GetConfiguration(base.Filepath(*testlog))
	if err != nil {
		log.Println("couldn't read configuration: ", err)
		return
	}

	fn, stype, suffix := input.Parse(flag.Arg(0))

	var entries []output.Entry

	if config.Connect {
		stime := time.Now()		
		entries, err = client.RemoteInvokeQuery(config, server.QueryBundle{fn, stype, suffix})
		if err != nil {
			log.Fatalln("problem connecting to server: ", err)
			return
		}
		log.Printf("query remote %v, %v, %v tool %v", fn, stype, suffix, time.Since(stime))		
	} else {
		gen := search.NewTrigramSearch(config.Indexpath, config.Prefixes)
		// TODO(rjk): error check
		entries, _ = gen.Query(fn, stype, []string{suffix})
	}
	output.WriteOut(os.Stdout, entries)
}
