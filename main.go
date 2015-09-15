package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/input"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
	"github.com/rjkroege/leap/server"
	"github.com/rjkroege/leap/client"
)

// TODO(rjk): It is conceivable that I will want to support having a re-writing
// path configuration: client can re-write paths for where the file is to be found
// in the client. I currently don't have this issue. So I'm not going to deal.
var (
	ip      = flag.Int("flagname", 1234, "help message for flagname")
	testlog = flag.Bool("testlog", false,
		"Log in the conventional way for running in a terminal. Also changes where to find the configuration file.")
	runServer = flag.Bool("server", false, "Run as a server. If a server is already running, does nothing.")
	stop = flag.Bool("stop", false, "Connect to the configured server and stop it.")
	index = flag.Bool("index", false, "Connect to the configured server and ask it to re-index the configured path.")
	host = flag.String("host", "", "Configure hostname for server. Empty host is short-circuited to operate in-memory.")
	indexpath = flag.String("indexpath", "",
		"Configure the path to the index file. Use CSEARCHINDEX if not provided. Client-only invocations ignore the configured index path.")
	resetpath = flag.Bool("resetpath", false,
		"Clear the configured index path.")
	remote = flag.Bool("remote", false,
		"Update the configuration file to specify that leap should operate in remote mode.")
	local = flag.Bool("local", false,
		"Update the configuration file to specify that leap should operate in local mode. Only one of -local and -remote can be specified.")
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

func updateConfigIfNecessary() {
	if !(*remote || *local || *host != "" || *indexpath != "" || *resetpath) {
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
	case *runServer:
		fmt.Fprintln(os.Stderr, "go run as server")
		config, err := base.GetConfiguration(base.Filepath(*testlog))
		if err != nil {
			log.Println("couldn't read configuration: ", err)
			return
		}
		server.BeginServing(config)
		os.Exit(0)
	case *stop:
		fmt.Fprintln(os.Stderr, "stop the running server")
		os.Exit(0)
	}

	// May exit.
	updateConfigIfNecessary()

	config, err := base.GetConfiguration(base.Filepath(*testlog))
	if err != nil {
		log.Println("couldn't read configuration: ", err)
		return
	}
	log.Println("config: ", *config)

	fn, stype, suffix := input.Parse(flag.Arg(0))

	var entries  []output.Entry

	if config.Connect {
		entries, err = client.RemoteInvokeQuery(config, server.QueryBundle{fn, stype, suffix})
		if err != nil {
			log.Fatalln("problem connecting to server: ", err);
		}
	} else {
		gen := search.NewTrigramSearch(config.Indexpath)
		// TODO(rjk): error check
		entries, _ = gen.Query(fn, stype, suffix)
	}
	output.WriteOut(os.Stdout, entries)
}
