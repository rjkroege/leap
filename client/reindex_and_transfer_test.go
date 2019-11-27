package client

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/index"
	"github.com/rjkroege/leap/server"
)

// Copied from the main line.
var (
	runServer = flag.Bool("server", false, "Run as a server. If a server is already running, does nothing.")

	// Need some additional things to stuff in the configuration.
	remoteindexfile = flag.String("remoteindexfile", "fail!", "Specify the path to the remote index file.")
)

// need to figure out how to launch the harness? It will be an instance of the
// test process. Edwood does this?

// TODO(rjk): There is considerable opportunity here to improve the code.
// The only state required for the server to run is the path fo the index
// path resident on the remote server. The configuration state is
// therefore unnecessary on the remote. I should simplify this. In
// particular, there is no benefit from passing around more state than
// necessary.
type MockServerConfiguration struct {
	newconfig      base.GlobalConfiguration
	originalconfig base.Configuration
}

func (msc *MockServerConfiguration) GetNewConfiguration() *base.GlobalConfiguration {
	return &msc.newconfig
}
func (msc *MockServerConfiguration) ClassicConfiguration() *base.Configuration {
	return &msc.originalconfig
}

func NewMockServerConfiguration(indexpath string) *MockServerConfiguration {
	return &MockServerConfiguration{
		//		newconfig: base.GlobalConfiguration {
		//			// TODO(rjk): Maybe some suff has to go here?
		//		},
		originalconfig: base.Configuration{
			Indexpath: indexpath,
		},
	}
}

type IntegrationTestDirectory struct {
	root              string
	remoteindexedpath string
	remoteindexfile   string
	localindexfile    string
}

func (itd *IntegrationTestDirectory) Cleanup(t *testing.T) {
	if err := os.RemoveAll(itd.root); err != nil {
		t.Errorf("Can't cleanup %s: %v", itd.root, err)
	}
}

func MakeIntegrationTestDirectory(t *testing.T) *IntegrationTestDirectory {
	pths := new(IntegrationTestDirectory)
	var err error

	pths.root, err = ioutil.TempDir("", "leap_integration_test")
	if err != nil {
		t.Fatalf("Can't create a test directory because %v", err)
	}

	pths.remoteindexedpath = filepath.Join(pths.root, "remote")
	if err := os.MkdirAll(pths.remoteindexedpath, 0755); err != nil {
		t.Fatalf("Can't make remoteindexedpath %s because %v", pths.remoteindexedpath, err)
	}

	pths.remoteindexfile = filepath.Join(pths.root, "remoteindex")
	pths.localindexfile = filepath.Join(pths.root, "localindex")

	inputs, err := filepath.Glob(filepath.Join("testdata", "*"))
	if err != nil {
		t.Fatalf("Can't enumerate test input data because %v", err)
	}
	for _, fn := range inputs {
		src, err := os.Open(fn)
		if err != nil {
			t.Fatalf("Can't open src file %s: %v", fn, err)
		}
		dstpath := filepath.Join(pths.remoteindexedpath, filepath.Base(fn))
		dst, err := os.Create(dstpath)
		if err != nil {
			t.Fatalf("Can't create dest file %s: %v", dstpath, err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			t.Fatalf("Can't copy %s to %s: %v", fn, dstpath, err)
		}
		src.Close()
		dst.Close()
	}

	logs, err := index.Idx{}.ReIndex(pths.remoteindexfile, pths.remoteindexedpath)
	if err != nil {
		t.Log(string(logs))
		t.Fatalf("Can't index %s: %v", pths.remoteindexedpath, err)
	}
	return pths
}

func TestMain(m *testing.M) {
	flag.Parse()

	if *runServer {
		log.Println("asked to run as a server")
		config := NewMockServerConfiguration(*remoteindexfile)
		server.BeginServing(config)
	} else {
		e := m.Run()
		os.Exit(e)
	}
}

// launchServerProcessHelper launches the test target as a server.
func launchServerProcessHelper(t *testing.T, args ...string) {
	leapserver := exec.Command(os.Args[0], args...)
	leapserver.Stdout = os.Stdout
	leapserver.Stderr = os.Stderr
	if err := leapserver.Start(); err != nil {
		t.Fatalf("failed to execute %s: %v", os.Args[0], err)
	}
}

// TestMakeTempState is about validating that we have wired up things
// properly to actually run more interesting tests. There is no need to
// run this after validating the result.
const onlyfordebugging = false

func TestMakeTempState(t *testing.T) {
	if onlyfordebugging {
		itd := MakeIntegrationTestDirectory(t)
		t.Logf("got an itd %v", itd)

		// Comment out to see the result.
		itd.Cleanup(t)
	}
}

// retries is the number of times to try connecting to the server.
const retries = 4

func tryConnecting() (*rpc.Client, error) {
	for i, delay := 0, 1; i < retries; i, delay = i+1, delay*10 {
		leapserver, err := rpc.DialHTTP("tcp", "localhost"+":1234")
		if err == nil {
			return leapserver, nil
		}
		timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
		<-timer.C
	}
	return nil, fmt.Errorf("too many retries without finding server")
}

// TestProcessLaunch forks a remote server and asks it a question. It
// largely serves to show that the test harness is working correctly.
func TestLaunchAndSayHello(t *testing.T) {
	itd := MakeIntegrationTestDirectory(t)
	defer itd.Cleanup(t)
	t.Logf("got an itd %v", itd)

	launchServerProcessHelper(t, "-server", "-remoteindexfile", itd.remoteindexfile)

	// There is a race condition here. There is no guarantee that the remote is up
	// yet.

	// TODO(rjk): Make sure that the port is configured?
	leapserver, err := tryConnecting()
	if err != nil {
		t.Fatalf("can't connect to remote: %s", err)
	}
	defer shutdownimpl(leapserver)

	var reply string
	err = leapserver.Call("Server.Ping", "Hõla!", &reply)
	if err != nil {
		t.Errorf("failed to message server: %v", err)
	}
	if got, want := reply, fmt.Sprintf("%s back to you!", "Hõla!"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// binarydiff returns true if a, b are the same bytes or error.
func fileequal(a, b string) (bool, error) {
	abs, err := ioutil.ReadFile(a)
	if err != nil {
		return false, fmt.Errorf("binarydiff can't open %s: %v", a, err)
	}
	bbs, err := ioutil.ReadFile(b)
	if err != nil {
		return false, fmt.Errorf("binarydiff can't open %s: %v", b, err)
	}
	return reflect.DeepEqual(abs, bbs), nil
}

const fourfile = `And now,
for something completely different
`

func (itd *IntegrationTestDirectory) insertFourFile() error {
	return ioutil.WriteFile(filepath.Join(itd.remoteindexedpath, "newfourfile"), []byte(fourfile), 0644)
}

// TestRemoteIndexAndTransfer exercises the full leap remote protocol and validates the result.
func TestRemoteIndexAndQuery(t *testing.T) {
	itd := MakeIntegrationTestDirectory(t)
	defer itd.Cleanup(t)

	launchServerProcessHelper(t, "-server", "-remoteindexfile", itd.remoteindexfile)
	leapserver, err := tryConnecting()
	if err != nil {
		t.Fatalf("can't connect to remote: %s", err)
	}
	defer shutdownimpl(leapserver)

	if err := reIndexAndTransferImpl(leapserver, itd.localindexfile, itd.remoteindexfile); err != nil {
		t.Errorf("reIndexAndTransferImpl failed: %v", err)
	}

	// Validate that the remote and local index files are the same
	switch equal, err := fileequal(itd.localindexfile, itd.remoteindexfile); {
	case err != nil:
		t.Errorf("can't compare files %s and %s: %v", itd.localindexfile, itd.remoteindexfile, err)
	case err == nil && !equal:
		t.Errorf("files %s and %s weren't equal", itd.localindexfile, itd.remoteindexfile)
	}

	// Search inside the result.
	leapIntegrationTestBeforeReindex(t, leapserver, itd)

	// Mutate indexed content (add a four file)
	if err := itd.insertFourFile(); err != nil {
		t.Errorf("can't add more data to 'remote' tree: %v", err)
	}

	if err := reIndexAndTransferImpl(leapserver, itd.localindexfile, itd.remoteindexfile); err != nil {
		t.Errorf("reIndexAndTransferImpl failed: %v", err)
	}

	// Validate that the remote and local index files are the same
	switch equal, err := fileequal(itd.localindexfile, itd.remoteindexfile); {
	case err != nil:
		t.Errorf("can't compare files %s and %s: %v", itd.localindexfile, itd.remoteindexfile, err)
	case err == nil && !equal:
		t.Errorf("files %s and %s weren't equal", itd.localindexfile, itd.remoteindexfile)
	}

	// Search inside the result.
	leapIntegrationTestAfterReindex(t, leapserver, itd)

	// I should be able to do name searches afterwards.
	shutdownimpl(leapserver)

	// Search inside the result.
	leapIntegrationTestAfterShuttingDownServer(t, leapserver, itd)

}
