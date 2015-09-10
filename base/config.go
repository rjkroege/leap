package base

import (
	"encoding/json"
	"path/filepath"
	"log"
	"os"
	"runtime"
)

// TODO(rjk): Do I need the attributes?
type Configuration struct {
	Hostname string
	Indexpath string
	Connect bool
}

func Filepath(test bool) string {
	if test {
		return "./leaprc"
	}

	var home string
	home = os.Getenv("HOME")
	if runtime.GOOS == "windows" && home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return filepath.Clean(home + "/.leaprc")
}

func GetConfiguration(fp string) (*Configuration, error) {
	log.Println("GetConfiguration")
	fd, err := os.Open(fp)

	if os.IsNotExist(err) {
		return &Configuration{"", "", false}, nil
	} else if err != nil {
		return nil, err
	}
	defer fd.Close()

	decoder := json.NewDecoder(fd)
	config := new(Configuration)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

// TODO(rjk): I'm not going to worry about simultaneous mutation.
func SaveConfiguration(config *Configuration, fp string) error {
	log.Println("SaveConfiguration")
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

func UpdateConfigurationFromCommandLine(fp, host, indexpath string, connect bool) error {
	config, err := GetConfiguration(fp)
	if err != nil {
		return err 
	}
	
	if host != "" {
		config.Hostname = host
	}
	if indexpath != "" {
		config.Indexpath = indexpath
	}
	config.Connect = connect
	return SaveConfiguration(config, fp)
}


