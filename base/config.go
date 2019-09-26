package base

import (
	"encoding/json"
	"fmt"
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
}

// The the new types.
type Project struct {
	host string  		`json:"host"`
	indexpath string	`json:"indexpath"`
	remote bool		`json:"remote"`
	prefixes []string	`json:"prefixes"`
}

type GobalConfiguration struct {
	version int			`json:"version"`
	currentproject string			`json:"currentproject"`
	projects map[string]Project  	`json:"projects"`
}

// getLegacyConfiguration returns the legacy Configuration object corresponding
// to the current project in the new style configuration.
func (gc *GobalConfiguration) getLegacyConfiguration() (*Configuration, error) {
	if _, ok := gc.projects[gc.currentproject] ; ok {
		return nil, fmt.Errorf("no project corresponding to selected project %s", gc.currentproject)
	}
	
	np := gc.projects[gc.currentproject]
	return &Configuration{
		Hostname:	np.host,
		Indexpath:	np.indexpath,
		Connect:		np.remote,
		Prefixes:		np.prefixes,
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

	decoder := json.NewDecoder(fd)
	newstyleconfig := new(GobalConfiguration)
	if err := decoder.Decode(newstyleconfig); err == nil && newstyleconfig.version == 1 {
		return newstyleconfig.getLegacyConfiguration()
	}

	if _, err := fd.Seek(0,0); err != nil {
		return nil, fmt.Errorf("can't seek on config file %s because %v", fp, err)
	}

	// Try decoding the old way.
	config := new(Configuration)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	if config.Indexpath == "" {
		config.Indexpath = index.File()
	}
	return config, nil
}

// TODO(rjk): I'm not going to worry about simultaneous mutation.
func SaveConfiguration(config *Configuration, fp string) error {
	fd, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer fd.Close()

	coder := json.NewEncoder(fd)
	if err := coder.Encode(config); err != nil {
		return err
	}
	return nil
}
