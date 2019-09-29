package base

import (
	"io/ioutil"
	"os"
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
		{
			"leaprc_new",
			Configuration{
				Hostname:  "myhost",
				Indexpath: "/home/gopher",
				Connect:   false,
				Prefixes:  []string{"/home/gopher/src"},
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

		econf := p.config
		if got, want := conf.Hostname, econf.Hostname; got != want {
			t.Errorf("%s wrong got %v want %v", "Hostname", got, want)
		}
		if got, want := conf.Indexpath, econf.Indexpath; got != want {
			t.Errorf("%s wrong got %v want %v", "Indexpath", got, want)
		}
		if got, want := conf.Connect, econf.Connect; got != want {
			t.Errorf("%s wrong got %v want %v", "Connect", got, want)
		}
		if got, want := conf.Prefixes, econf.Prefixes; !reflect.DeepEqual(got, want) {
			t.Errorf("%s wrong got %v want %v", "Prefixes", got, want)
		}
	}
}

func TestSaveNewConfig(t *testing.T) {
	fp := filepath.Join("testdata", "leaprc_original")
	oconf, err := GetConfiguration(fp)

	nconf := updateConfig(oconf)

	// TODO(rjk): make temp file
	fd, err := ioutil.TempFile("", "leap_config")
	if err != nil {
		t.Fatalf("can't make a temp file because %v", err)
	}
	fd.Close()
	name := fd.Name()
	defer os.Remove(name)

	if err := saveNewConfig(nconf, name); err != nil {
		t.Fatalf("can't write a config because %v", err)
	}

	// Should be able to get legacy from upgraded.
	conf, err := GetConfiguration(name)
	if err != nil {
		t.Fatalf("can't read a new-style config because %v", err)
	}

	econf := Configuration{
		Hostname:  "",
		Indexpath: "/Users/rjkroege/.csearchindex",
		Connect:   false,
		Prefixes:  []string{"/Users/rjkroege/tools/gopkg/src"},
	}

	if got, want := conf.Hostname, econf.Hostname; got != want {
		t.Errorf("%s wrong got %v want %v", "Hostname", got, want)
	}
	if got, want := conf.Indexpath, econf.Indexpath; got != want {
		t.Errorf("%s wrong got %v want %v", "Indexpath", got, want)
	}
	if got, want := conf.Connect, econf.Connect; got != want {
		t.Errorf("%s wrong got %v want %v", "Connect", got, want)
	}
	if got, want := conf.Prefixes, econf.Prefixes; !reflect.DeepEqual(got, want) {
		t.Errorf("%s wrong got %v want %v", "Prefixes", got, want)
	}
}

func TestUpdateNewConfig(t *testing.T) {
	fd, err := ioutil.TempFile("", "leap_config")
	if err != nil {
		t.Fatalf("can't make a temp file because %v", err)
	}
	name := fd.Name()
	fd.Close()
	defer os.Remove(name)

	// Copy leaprc_new to temp file
	fp := filepath.Join("testdata", "leaprc_new")
	contents, err := ioutil.ReadFile(fp)
	if err != nil {
		t.Fatalf("can't read starter config because %v", err)
	}
	if err := ioutil.WriteFile(name, contents, 0644); err != nil {
		t.Fatalf("can't read starter config because %v", err)
	}

	oconf, err := GetConfiguration(name)
	if err != nil {
		t.Fatalf("can't read starter config because %v", err)
	}

	oconf.Hostname = "pinkelephant"
	oconf.Indexpath = "/home/augie"
	oconf.Connect = true
	oconf.Prefixes = []string{"/home/augie/go/src"}

	if err := SaveConfiguration(oconf, name); err != nil {
		t.Fatalf("can't write starter config because %v", err)
	}

	conf, err := GetConfiguration(name)
	if err != nil {
		t.Fatalf("can't read modified config because %v", err)
	}

	econf := Configuration{
		Hostname:  "pinkelephant",
		Indexpath: "/home/augie",
		Connect:   true,
		Prefixes:  []string{"/home/augie/go/src"},
	}

	if got, want := conf.Hostname, econf.Hostname; got != want {
		t.Errorf("%s wrong got %v want %v", "Hostname", got, want)
	}
	if got, want := conf.Indexpath, econf.Indexpath; got != want {
		t.Errorf("%s wrong got %v want %v", "Indexpath", got, want)
	}
	if got, want := conf.Connect, econf.Connect; got != want {
		t.Errorf("%s wrong got %v want %v", "Connect", got, want)
	}
	if got, want := conf.Prefixes, econf.Prefixes; !reflect.DeepEqual(got, want) {
		t.Errorf("%s wrong got %v want %v", "Prefixes", got, want)
	}
}
