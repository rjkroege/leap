package base

import (
	"encoding/json"
	"path/filepath"
	"os"
	"runtime"

	"github.com/google/codesearch/index"
)

// See package comment for why we have this amusing string.
const Prefix = "/tmp/.leaping/glenda"
const SubPrefix = "/tmp/.leaping"

// TODO(rjk): Do I need the attributes?
type Configuration struct {
	Hostname string
	Indexpath string
	Connect bool
	Prefixes []string
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
	defer fd.Close()

	decoder := json.NewDecoder(fd)
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
