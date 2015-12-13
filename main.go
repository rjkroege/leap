package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/client"
	"github.com/rjkroege/leap/highlights"
	"github.com/rjkroege/leap/input"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
	"github.com/rjkroege/leap/server"
)

// TODO(rjk): It is conceivable that I will want to support having a re-writing
// path configuration: client can re-write paths for where the file is to be found
// in the client. I currently don't have this issue. So I'm not going to deal.
var (
	ip      = flag.Int("flagname", 1234, "help message for flagname")
	testlog = flag.Bool("testlog", false,
		"Log in the conventional way for running in a terminal. Also changes where to find the configuration file.")
	runServer = flag.Bool("server", false, "Run as a server. If a server is already running, does nothing.")
	stop      = flag.Bool("stop", false, "Connect to the configured server and stop it.")
	index     = flag.Bool("index", false, "Connect to the configured server and ask it to re-index the configured path.")
	host      = flag.String("host", "", "Configure hostname for server. Empty host is short-circuited to operate in-memory.")
	indexpath = flag.String("indexpath", "",
		"Configure the path to the index file. Use CSEARCHINDEX if not provided. Client-only invocations ignore the configured index path.")
	resetpath = flag.Bool("resetpath", false,
		"Clear the configured index path.")
	remote = flag.Bool("remote", false,
		"Update the configuration file to specify that leap should operate in remote mode.")
	local = flag.Bool("local", false,
		"Update the configuration file to specify that leap should operate in local mode. Only one of -local and -remote can be specified.")
	setprefix = flag.Bool("setprefix", false,
		"Set the path trimming prefixes to the given paths.")
	decodePlumb = flag.Bool("dp", false,
		"Decode the single provided path and convert it back into a valid plumb address")
	highlightFile = flag.Bool("hf", false,
		"Reprocess the output of highlight marking the line extracted from the provided encoded path")
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

func updateConfigIfNecessary(args []string) {
	if !(*remote || *local || *host != "" || *indexpath != "" || *resetpath || *setprefix) {
		return
	}

	fp := base.Filepath(*testlog)
	config, err := base.GetConfiguration(fp)
	if err != nil {
		log.Println("Failed to read configuration: ", err)
	}

	switch {
	case *remote && *local:
		flag.Usage()
		os.Exit(1)
	case *remote:
		config.Connect = true
	case *local:
		config.Connect = false
	}

	switch {
	case *resetpath && *indexpath != "":
		flag.Usage()
		os.Exit(1)
	case *resetpath:
		config.Indexpath = ""
	case *indexpath != "":
		config.Indexpath = *indexpath
	}

	if *host != "" {
		config.Hostname = *host
	}

	if *setprefix {
		log.Println("Setprefix", args)
		config.Prefixes = args
	}

	if err := base.SaveConfiguration(config, fp); err != nil {
		log.Fatalf("Failed to write configuration: ", err)
	}
	os.Exit(0)
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
	case *highlightFile:
		log.Println("running highlightFile", flag.Arg(0))
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(0)
		}
		if err := highlights.ShowDesiredLineInFile(input.EncodedToNumber(flag.Arg(0)), os.Stdin, os.Stdout); err != nil {
			log.Println("ShowDesiredLine... failed ", err)
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
	updateConfigIfNecessary(flag.Args())
	os.RemoveAll(base.SubPrefix)

	config, err := base.GetConfiguration(base.Filepath(*testlog))
	if err != nil {
		log.Println("couldn't read configuration: ", err)
		return
	}
	log.Println("config: ", *config)

	fn, stype, suffix := input.Parse(flag.Arg(0))

	var entries []output.Entry

	if config.Connect {
		entries, err = client.RemoteInvokeQuery(config, server.QueryBundle{fn, stype, suffix})
		if err != nil {
			log.Fatalln("problem connecting to server: ", err)
		}
	} else {
		gen := search.NewTrigramSearch(config.Indexpath, config.Prefixes)
		// TODO(rjk): error check
		entries, _ = gen.Query(fn, stype, []string{suffix})
	}
	output.WriteOut(os.Stdout, entries)
}
