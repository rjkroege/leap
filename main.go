package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/client"
	"github.com/rjkroege/leap/index"
	"github.com/rjkroege/leap/input"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
	"github.com/rjkroege/leap/server"
	// Uncomment to turn on profiling.
	// "github.com/pkg/profile"
	"github.com/rjkroege/gozen"
)

// TODO(rjk): It is conceivable that I will want to support having a re-writing
// path configuration: client can re-write paths for where the file is to be found
// in the client. I currently don't have this issue. So I'm not going to deal.
var (
	testlog = flag.Bool("testlog", false,
		"Log in the conventional way for running in a terminal. Also changes where to find the configuration file.")
	runServer = flag.Bool("server", false, "Run as a server. If a server is already running, does nothing.")
	stop      = flag.Bool("stop", false, "Connect to the configured server and stop it.")

	indexcmd    = flag.Bool("index", false, "Connect to the configured server and ask it to re-index the configured path.")
	decodePlumb = flag.Bool("dp", false,
		"Decode the single provided path and convert it back into a valid plumb address")

	printcsindex = flag.Bool("cspath", false, "Print the path needed for CSEARCHINDEX")
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
		if err := base.SetupLogging("leap"); err != nil {
			log.Fatalf("can't setup right logging: %v", err)
		}
		defer base.RollLogs("leap")
	}

	switch {
	case *decodePlumb:
		log.Println("running decodePlumb")
		os.RemoveAll(base.SubPrefix)
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(0)
		}
		path := input.EncodedToPlumb(flag.Arg(0))
		log.Println("output", path)
		// fmt.Println(path)
		if err := gozen.Editinacme(path); err != nil {
			log.Fatalf("can't tell Edwood/Acme to open %s: %v", path, err)
		}
		os.Exit(0)
	case *runServer:
		fmt.Fprintln(os.Stderr, "go run as server")
		config, err := base.GetConfiguration(base.Filepath(*testlog))
		if err != nil {
			log.Fatal("couldn't read configuration: ", err)
		}
		server.BeginServing(config)
		os.Exit(0)
	case *printcsindex:
		config, err := base.GetConfiguration(base.Filepath(*testlog))
		if err != nil {
			log.Fatal("couldn't read configuration: ", err)
		}
		fmt.Printf(config.Indexpath)
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
	case *indexcmd:
		// TODO(rjk): Pull this block out into a helper function.
		config, err := base.GetConfiguration(base.Filepath(*testlog))
		if err != nil {
			log.Fatal("couldn't read configuration: ", err)
			return
		}
		newconfig := config.GetNewConfiguration()
		if newconfig == nil {
			log.Fatal("index command requires upgraded config")
			return
		}

		if config.Connect {
			if err := client.ReIndexAndTransfer(newconfig); err != nil {
				log.Println("Remote index failed because: ", err)
			}
		} else {
			// TODO(rjk): I can probably make this prettier.
			output, err := index.Idx{}.ReIndex(newconfig.Projects[newconfig.Currentproject].Remotepath, newconfig.Currentproject)
			if err != nil {
				fmt.Printf("couldn't reindex because: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
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

	if config.Connect && stype != ":" {
		stime := time.Now()
		search := search.NewTrigramSearch(config.Indexpath, config.Prefixes)
		log.Printf("running remote query after NewTrigramSearch %v\n", time.Since(stime))
		// TODO(rjk): Dialing the remote can be expensive because ssh. I should overlap
		// the connect with the search of the local index.
		inremotes, err := client.NewRemoteInternalSearcher(config)
		if err != nil {
			log.Fatalln("problem connecting to server: ", err)
			return
		}
		log.Printf("running remote query after NewRemoteInternalSearcher %v\n", time.Since(stime))
		entries, err = search.Query(fn, stype, []string{suffix}, inremotes)
		log.Printf("query remote %v, %v, %v tool %v\n", fn, stype, suffix, time.Since(stime))
	} else {
		stime := time.Now()
		search := search.NewTrigramSearch(config.Indexpath, config.Prefixes)
		// TODO(rjk): error check
		entries, _ = search.Query(fn, stype, []string{suffix}, search)
		log.Printf("query local %v, %v, %v tool %v\n", fn, stype, suffix, time.Since(stime))
	}
	stime := time.Now()
	output.WriteOut(os.Stdout, entries)
	log.Printf("after query, WriteOut %v\n", time.Since(stime))
}
