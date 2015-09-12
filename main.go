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
		"Configure the path to the index file. Use CSEARCHINDEX if not provided. Ignored by a client connecting to a server.")
	remote = flag.Bool("remote", false,
		"Update the configuration file to specify that leap should operate in remote mode.")
	local = flag.Bool("local", false,
		"Update the configuration file to specify that leap should operate in local mode.")
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

	if *runServer {
		fmt.Fprintln(os.Stderr, "go run as server")
		config, err := base.GetConfiguration(base.Filepath(*testlog))
		if err != nil {
			log.Println("couldn't read configuration: ", err)
			return
		}
		// I want this to stop here...
		server.BeginServing(config)
		return
	}
	if *stop {
		fmt.Fprintln(os.Stderr, "stop the running server")
		return
	}
	if *remote || *local || *host != "" || *indexpath != "" {
		connect := false
		if *remote {
			connect = true
		} else if *local {
			connect = false
		}
		if err := base.UpdateConfigurationFromCommandLine(base.Filepath(*testlog), *host, *indexpath, connect); err != nil {
			log.Println("failed to update configuration: ", err)
		}
		return
	}

	config, err := base.GetConfiguration(base.Filepath(*testlog))
	if err != nil {
		log.Println("couldn't read configuration: ", err)
		return
	}
	log.Println("config: ", *config)

	fn, stype, suffix := input.Parse(flag.Arg(0))
	gen := search.NewTrigramSearch()

	// TODO(rjk): error check
	entries, _ := gen.Query(fn, stype, suffix)
	output.WriteOut(os.Stdout, entries)
}
