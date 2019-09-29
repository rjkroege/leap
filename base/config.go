package base

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"os"
	"runtime"

	"github.com/google/codesearch/index"
)

// See package comment for why we have this amusing string.
const Prefix = "/tmp/.leaping/glenda"
const SubPrefix = "/tmp/.leaping"

type Configuration struct {
	Hostname string
	Indexpath string
	Connect bool
	Prefixes []string
	newconfig *GlobalConfiguration
}

// The the new types.
type Project struct {
	Host string  		`json:"host"`
	Indexpath string	`json:"indexpath"`
	Remote bool		`json:"remote"`
	Prefixes []string	`json:"prefixes"`
}

type GlobalConfiguration struct {
	Version int			`json:"version"`
	Currentproject string			`json:"currentproject"`
	Projects map[string]*Project  	`json:"projects"`
}

// getLegacyConfiguration returns the legacy Configuration object corresponding
// to the current project in the new style configuration.
func (gc *GlobalConfiguration) getLegacyConfiguration() (*Configuration, error) {
	if _, ok := gc.Projects[gc.Currentproject] ; !ok {
		return nil, fmt.Errorf("no project corresponding to selected project %s", gc.Currentproject)
	}
	
	np := gc.Projects[gc.Currentproject]
	return &Configuration{
		Hostname:	np.Host,
		Indexpath:	np.Indexpath,
		Connect:		np.Remote,
		Prefixes:		np.Prefixes,
		newconfig:	gc,
	}, nil
}


// Filepath returns the path to the leaprc configuration file.
func Filepath(test bool) string {
	if test {
		return "leaprc"
	}

	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" && home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return filepath.Clean(filepath.Join(home ,".leaprc"))
}

func GetConfiguration(fp string) (*Configuration, error) {
	fd, err := os.Open(fp)

	if os.IsNotExist(err) {
		return &Configuration{
			Hostname: "",
			Indexpath: index.File(),
			Connect: false,
		}, nil
	} else if err != nil {
		return nil, err
	}

	if ns, err := getNewConfig(fd); err == nil {
		a, b := ns.getLegacyConfiguration()
		return a, b
	}

	// New way didn't work so try decoding the old way.
	if _, err := fd.Seek(0,0); err != nil {
		return nil, fmt.Errorf("can't seek on config file %s because %v", fp, err)
	}

	config := new(Configuration)
	decoder := json.NewDecoder(fd)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	if config.Indexpath == "" {
		config.Indexpath = index.File()
	}
	return config, nil
}

// getNewConfig reads the new style configuration data from disk.
func getNewConfig(reader io.Reader) (*GlobalConfiguration, error) {
	decoder := json.NewDecoder(reader)
	newstyleconfig := new(GlobalConfiguration)
	if err := decoder.Decode(newstyleconfig); err != nil {
		return nil, err
	}
	if newstyleconfig.Version != 1 {
		return nil, fmt.Errorf("unsupported newconfig version, can't decode")
	}
	return newstyleconfig, nil
}

// updateConfig makes a new-style configuration out of the old one.
func updateConfig(oldconfig *Configuration) *GlobalConfiguration {
	return &GlobalConfiguration{
		Version: 1,
		Currentproject: "default",
		Projects: map[string]*Project {
			"default": &Project{
				Host: oldconfig.Hostname,
				Indexpath: oldconfig.Indexpath,
				Remote: oldconfig.Connect,
				Prefixes: oldconfig.Prefixes,
			},
		},
	}
}

// saveNewConfig writes a new config to disk or generates an error.
func saveNewConfig(config *GlobalConfiguration, fp string) error {
	fd, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer fd.Close()

	coder := json.NewEncoder(fd)
	coder.SetIndent("", "	")
	if err := coder.Encode(config); err != nil {
		return err
	}
	return nil
}

// updateNewConfiguration will update config's embedded newconfig based
// on changes made to the legacy Configuration payloads.
func (config *Configuration) pushConfigIntoNew() {
	nc := config.newconfig
	
	proj := nc.Projects[nc.Currentproject]

	proj.Host = config.Hostname
	proj.Indexpath = config.Indexpath
	proj.Remote = config.Connect
	proj.Prefixes = config.Prefixes
}

// TODO(rjk): I'm not going to worry about simultaneous mutation.
// This code does not force conversion to the new format. I want
// to do this manually via a top-level leap command to avoid
// surprising myself.
func SaveConfiguration(config *Configuration, fp string) error {
	// We might have a new style config. If we do, we update that
	// instead. Otherwise, we write out the old way.

	fd, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("can't open file for config because %v", err)
	}
	defer fd.Close()

	if _, err := getNewConfig(fd); err == nil {
		// We have a new configuration so update that instead.
		config.pushConfigIntoNew()
		return saveNewConfig(config.newconfig, fp)
	}

	// We a old configuration. So update that.
	coder := json.NewEncoder(fd)
	if err := coder.Encode(config); err != nil {
		return err
	}
	return nil
}
