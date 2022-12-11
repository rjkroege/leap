package base

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	update      = flag.Bool("updateconfig", false, "Upgrade the configuration version to the new format. Must be used by itself.")
	setproject  = flag.String("proj", "", "Set the current project configuration or create a new one")
	listproject = flag.Bool("lsproj", false, "List the projects")
)

// updateNewStyleConfiguration updates ncc based on the args. We can't
// get here without actually already having a new-style config. update
// must be the only command and needs to be handled separately.
func (ncc *GlobalConfiguration) updateNewStyleConfiguration(path string, args []string) {
	// Exit if we are an unsupported command combination.
	if *update || *remote && *local || *listproject && *setproject != "" || *resetpath && *indexpath != "" {
		fmt.Println("Invalid command combination.")
		flag.Usage()
		os.Exit(1)
	}

	if *listproject {
		// TODO(rjk): Dump the necessary content to give me autoocomplete list
		// The goal is to make this auto-complete  capable in Alfred
		s := make([]string, len(ncc.Projects))
		for k := range ncc.Projects {
			s = append(s, k)
		}
		fmt.Println(strings.Join(s, ", "))
		os.Exit(1)
	}

	if *setproject != "" {
		if _, ok := ncc.Projects[*setproject]; !ok {
			ncc.Projects[*setproject] = &Project{
				Host:          "",
				Indexpath:     "",
				Remote:        false,
				Prefixes:      []string{},
				Remoteproject: "",
				Remotepath:    "",
			}
		}
		ncc.Currentproject = *setproject
	}
	cp := ncc.Projects[ncc.Currentproject]

	switch {
	case *remote:
		cp.Remote = true
	case *local:
		cp.Remote = false
	}

	switch {
	case *resetpath:
		cp.Indexpath = ""
	case *indexpath != "":
		cp.Indexpath = *indexpath
	}

	if *host != "" {
		cp.Indexpath = *indexpath
	}

	if *setprefix {
		cp.Prefixes = args
	}

	if err := saveNewConfig(ncc, path); err != nil {
		log.Fatalf("Failed to write configuration: %v", err)
	}
	os.Exit(0)
}

// UpdateConfigIfNecessary adjusts the saved configuration based on
// command line flags. May exit the program.
// TODO(rjk): This is not well structured. This code could be better.
func UpdateConfigIfNecessary(args []string, testingconfig bool) {
	if !(*remote || *local || *host != "" || *indexpath != "" || *resetpath || *setprefix || *update || *listproject || *setproject != "") {
		return
	}

	fp := Filepath(testingconfig)
	fd, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Failed to open configuration %s because %v\n", fp, err)
	}

	newconfig, err := getNewConfig(fd)
	if err != nil {
		fmt.Printf("Can't treat %s as new configuration because %v\n", fp, err)
	}

	if newconfig != nil {
		// We have a new configuration.
		newconfig.updateNewStyleConfiguration(fp, args)
		return
	}

	config, err := GetConfiguration(fp)
	if *update {
		newconfig := updateConfig(config)
		if err := saveNewConfig(newconfig, fp); err != nil {
			log.Printf("can't update config %s because %v", fp, err)
		}
		os.Exit(1)
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
		log.Fatalf("Failed to write configuration: %v", err)
	}
	os.Exit(0)
}
