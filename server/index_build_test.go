package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/Redundancy/go-sync/chunks"
	"github.com/Redundancy/go-sync/filechecksum"
	grsync "github.com/Redundancy/go-sync/index"
	"github.com/rjkroege/leap/base"
	"github.com/sanity-io/litter"
)

// TODO(rjk): It might be a good idea to pull out the type of the test vector and put
// the functions on it?
type MockFailedConfiguration struct{}

func (_ MockFailedConfiguration) GetNewConfiguration() *base.GlobalConfiguration { return nil }
func (_ MockFailedConfiguration) ClassicConfiguration() *base.Configuration      { return nil }

type MockIndexer struct {
	result []byte
	err    error
	proj   string
}

func (mi MockIndexer) ReIndex(config *base.GlobalConfiguration, currentproject string) ([]byte, error) {
	if currentproject != mi.proj {
		return nil, fmt.Errorf("expected currentproject %s, got %s", mi.proj, currentproject)
	}
	return mi.result, mi.err
}

// TODO(rjk): need a more useful mock.
type MockSuccessConfiguration struct{}

func (_ MockSuccessConfiguration) GetNewConfiguration() *base.GlobalConfiguration {
	return &base.GlobalConfiguration{
		Version:        1,
		Currentproject: "fooproj",
		// TODO(rjk): This is more that goes here.
	}
}
func (_ MockSuccessConfiguration) ClassicConfiguration() *base.Configuration { return nil }

// MockFilesystemPostStatError implements the filesystem interface in a
// way that experiences an error after a successful stat. (Note that
// filename has to exist.)
type MockFilesystemPostStatError struct{}

func (_ MockFilesystemPostStatError) Stat(filename string) (os.FileInfo, error) {
	fi, err := os.Stat(filename)
	os.Remove(filename)
	return fi, err
}

// MockFailingBuilder is a builder implementation that fails.
type MockFailingBuilder struct{}

func (_ MockFailingBuilder) BuildChecksumIndex(check *filechecksum.FileChecksumGenerator, r io.Reader) (
	[]byte,
	*grsync.ChecksumIndex,
	filechecksum.ChecksumLookup,
	error,
) {
	return nil, nil, nil, fmt.Errorf("bad BuildChecksumIndex")
}

type MockBadChecksumLookup struct{}

func (_ MockBadChecksumLookup) GetStrongChecksumForBlock(_ int) []byte {
	return nil
}

type MockBadBuildChecksumIndexResults struct{}

func (_ MockBadBuildChecksumIndexResults) BuildChecksumIndex(check *filechecksum.FileChecksumGenerator, r io.Reader) (
	[]byte,
	*grsync.ChecksumIndex,
	filechecksum.ChecksumLookup,
	error,
) {
	return nil, nil, MockBadChecksumLookup{}, nil
}

type testVector struct {
	name          string
	server        Server
	err           error
	args          IndexAndBuildChecksumIndexArgs
	got           chunks.StrongChecksumGetter
	expectedtoken int
	pretest       func(t *testing.T, tv *testVector)
	posttest      func(t *testing.T, tv *testVector)
}

func TestIndexAndBuildChecksumIndex(t *testing.T) {
	tests := []testVector{
		{
			name: "bad token test",
			server: Server{
				token: 2,
			},
			err: fmt.Errorf("token mis-match: two syncs in progress?"),
			args: IndexAndBuildChecksumIndexArgs{
				Token: 1,
			},
			expectedtoken: 2,
		},
		{
			name: "bad config test",
			server: Server{
				token:  0,
				config: MockFailedConfiguration{},
			},
			err: fmt.Errorf("index command requires upgrading config"),
			args: IndexAndBuildChecksumIndexArgs{
				Token:             1,
				RemoteProjectName: "fooproj",
				RemotePath:        "foopath",
			},
			expectedtoken: 0,
		},
		{
			name: "cindex failed test",
			server: Server{
				token:  0,
				config: MockSuccessConfiguration{},
				indexer: MockIndexer{
					result: []byte{},
					err:    fmt.Errorf("can't fork cindex!"),
					proj:   "fooproj",
				},
			},
			err: fmt.Errorf("remote index command failed because: can't fork cindex!"),
			args: IndexAndBuildChecksumIndexArgs{
				Token:             1,
				RemoteProjectName: "fooproj",
				RemotePath:        "foopath",
			},
			expectedtoken: 0,
		},
		{
			name: "cindex success but filepath not available",
			server: Server{
				token:  0,
				config: MockSuccessConfiguration{},
				indexer: MockIndexer{
					result: []byte("cindex success!"),
					err:    nil,
					proj:   "fooproj",
				},
				fs: MockFilesystemPostStatError{},
			},
			err: fmt.Errorf("can't stat remote index foopath because stat foopath: no such file or directory"),
			args: IndexAndBuildChecksumIndexArgs{
				Token:             1,
				RemoteProjectName: "fooproj",
				RemotePath:        "foopath",
			},
			expectedtoken: 0,
		},
		{
			name: "stat succeeded, open fails",
			server: Server{
				token:  0,
				config: MockSuccessConfiguration{},
				indexer: MockIndexer{
					result: []byte("cindex success!"),
					err:    nil,
					proj:   "fooproj",
				},
				fs: MockFilesystemPostStatError{},
			},
			args: IndexAndBuildChecksumIndexArgs{
				Token:             1,
				RemoteProjectName: "fooproj",
			},
			expectedtoken: 0,
			pretest: func(t *testing.T, tv *testVector) {
				// make a temporary file.
				fd, err := ioutil.TempFile("", "index_build_test_stat_success")
				if err != nil {
					t.Fatalf("Can't make temp file for %s: %v", tv.name, err)
				}
				if _, err := io.WriteString(fd, "hello there"); err != nil {
					t.Fatalf("Can't write data to temp file %s: %v", tv.name, err)
				}
				tv.args.RemotePath = fd.Name()
				tv.err = fmt.Errorf("can't open remote index %s because open %s: no such file or directory", fd.Name(), fd.Name())
				fd.Close()
			},
			posttest: func(t *testing.T, tv *testVector) {
				os.Remove(tv.args.RemotePath)
			},
		},
		{
			name: "BuildCheckSum fails",
			server: Server{
				token:  0,
				config: MockSuccessConfiguration{},
				indexer: MockIndexer{
					result: []byte("cindex success!"),
					err:    nil,
					proj:   "fooproj",
				},
				fs:    filesystemimpl{},
				build: MockFailingBuilder{},
			},
			args: IndexAndBuildChecksumIndexArgs{
				Token:             1,
				RemoteProjectName: "fooproj",
			},
			expectedtoken: 0,
			pretest: func(t *testing.T, tv *testVector) {
				// make a temporary file.
				fd, err := ioutil.TempFile("", "index_build_test_buildchecksum_fails")
				if err != nil {
					t.Fatalf("Can't make temp file for %s: %v", tv.name, err)
				}
				if _, err := io.WriteString(fd, "hello there"); err != nil {
					t.Fatalf("Can't write data to temp file %s: %v", tv.name, err)
				}
				tv.args.RemotePath = fd.Name()
				tv.err = fmt.Errorf("can't compute checksums on %s because bad BuildChecksumIndex", fd.Name())
				fd.Close()
			},
			posttest: func(t *testing.T, tv *testVector) {
				if err := tv.server.indexfile.Close(); err == nil {
					t.Errorf("server.indexfile hadn't been closed")
				}
				os.Remove(tv.args.RemotePath)
			},
		},
		{
			name: "BuildCheckSum generates invalid results",
			server: Server{
				token:  0,
				config: MockSuccessConfiguration{},
				indexer: MockIndexer{
					result: []byte("cindex success!"),
					err:    nil,
					proj:   "fooproj",
				},
				fs:    filesystemimpl{},
				build: MockBadBuildChecksumIndexResults{},
			},
			args: IndexAndBuildChecksumIndexArgs{
				Token:             1,
				RemoteProjectName: "fooproj",
			},
			expectedtoken: 0,
			err:           fmt.Errorf("can't convert checksumLookup into a concrete StrongChecksumGetter"),
			pretest: func(t *testing.T, tv *testVector) {
				// make a temporary file.
				fd, err := ioutil.TempFile("", "index_build_test_buildchecksum_fails")
				if err != nil {
					t.Fatalf("Can't make temp file for %s: %v", tv.name, err)
				}
				if _, err := io.WriteString(fd, "hello there"); err != nil {
					t.Fatalf("Can't write data to temp file %s: %v", tv.name, err)
				}
				tv.args.RemotePath = fd.Name()
				fd.Close()
			},
			posttest: func(t *testing.T, tv *testVector) {
				if err := tv.server.indexfile.Close(); err == nil {
					t.Errorf("server.indexfile hadn't been closed")
				}
				os.Remove(tv.args.RemotePath)
			},
		},
		{
			name: "BuildCheckSum returned correctly",
			server: Server{
				token:  0,
				config: MockSuccessConfiguration{},
				indexer: MockIndexer{
					result: []byte("cindex success!"),
					err:    nil,
					proj:   "fooproj",
				},
				fs:    filesystemimpl{},
				build: builderimpl{},
			},
			args: IndexAndBuildChecksumIndexArgs{
				Token:             1,
				RemoteProjectName: "fooproj",
			},
			expectedtoken: 1,
			err:           nil,
			got: chunks.StrongChecksumGetter{
				chunks.ChunkChecksum{
					ChunkOffset: 0,
					Size:        0,
					WeakChecksum: []uint8{
						76,
						4,
						187,
						25,
					},
					StrongChecksum: []uint8{
						22,
						27,
						194,
						89,
						98,
						218,
						143,
						237,
						109,
						47,
						89,
						146,
						47,
						182,
						66,
						170,
					},
				},
			},
			pretest: func(t *testing.T, tv *testVector) {
				// make a temporary file.
				fd, err := ioutil.TempFile("", "index_build_test_buildchecksum_fails")
				if err != nil {
					t.Fatalf("Can't make temp file for %s: %v", tv.name, err)
				}
				if _, err := io.WriteString(fd, "hello there"); err != nil {
					t.Fatalf("Can't write data to temp file %s: %v", tv.name, err)
				}
				tv.args.RemotePath = fd.Name()
				fd.Close()
			},
			posttest: func(t *testing.T, tv *testVector) {
				if err := tv.server.indexfile.Close(); err != nil {
					t.Errorf("server.indexfile should have been open but wasn't")
				}
				os.Remove(tv.args.RemotePath)
			},
		},
	}

	for _, test := range tests {
		if test.pretest != nil {
			test.pretest(t, &test)
		}

		var result RemoteCheckSumIndexData
		err := test.server.IndexAndBuildChecksumIndex(test.args, &result)

		if test.err != nil && err == nil {
			t.Errorf("%s expected error %v but got no error", test.name, test.err)
		} else if test.err != nil && err != nil {
			if got, want := err.Error(), test.err.Error(); got != want {
				t.Errorf("%s expected error %v, got error %v", test.name, test.err, err)
			}
		} else if test.err == nil && err != nil {
			t.Errorf("%s got unexpected error %v", test.name, err)
		} else {
			// We only look at the StrongChecksumGetter in the results because
			// it's derived from the rest and the checksum code is (hopefully) tested
			// elsewhere.
			if got, want := result.StrongChecksumGetter, test.got; !reflect.DeepEqual(got, want) {
				t.Errorf("%s got unexpected result. got %s, want %s", test.name, litter.Sdump(got), litter.Sdump(want))
			}
		}

		if got, want := test.server.token, test.expectedtoken; got != want {
			t.Errorf("%s did not correctly set token got %d want %d", test.name, got, want)
		}

		if test.posttest != nil {
			test.posttest(t, &test)
		}
	}
}
