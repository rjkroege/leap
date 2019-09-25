package base

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetConfiguration(t *testing.T) {
	tt := []struct {
		name   string
		config Configuration
	}{
		{
			"leaprc_original",
			Configuration{
				Hostname:  "",
				Indexpath: "/Users/rjkroege/.csearchindex",
				Connect:   false,
				Prefixes:  []string{"/Users/rjkroege/tools/gopkg/src"},
			},
		},
	}

	for _, p := range tt {
		fp := filepath.Join("testdata", p.name)

		conf, err := GetConfiguration(fp)
		if err != nil {
			t.Errorf("unexpected error %v for config %s", err, fp)
			continue
		}

		if !reflect.DeepEqual(p.config, *conf) {
			// TODO(rjk): use mods, use the pretty-printer, etc.
			t.Errorf("didn't get desired value for %s: got %v, want %v", fp, conf, p.config)
		}
	}
}

//func TestGetConfigurationMissing(t *testing.T) {
//}
