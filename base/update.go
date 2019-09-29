package base

import (
	"flag"
	"log"
	"os"
)

var (
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
)

// UpdateConfigIfNecessary adjusts the saved configuration based on
// command line flags. May exit the program.
func UpdateConfigIfNecessary(args []string, testingconfig bool) {
	if !(*remote || *local || *host != "" || *indexpath != "" || *resetpath || *setprefix) {
		return
	}

	fp := Filepath(testingconfig)
	config, err := GetConfiguration(fp)
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

	if err := SaveConfiguration(config, fp); err != nil {
		log.Fatalf("Failed to write configuration: ", err)
	}
	os.Exit(0)
}
