package client

import (
	"net/rpc"
//	"log"
	"fmt"
	"crypto/md5"
	"os"
	"path/filepath"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
	// "github.com/rjkroege/leap/search"
	"github.com/rjkroege/leap/server"
	//"github.com/codeskyblue/go-sh"
	"github.com/Redundancy/go-sync"
	"github.com/Redundancy/go-sync/blocksources"
	"github.com/Redundancy/go-sync/filechecksum"
)

const (
	MB = 1024 * 1024
)


// TODO(rjk): Make the port configurable?
func RemoteInvokeQuery(config *base.Configuration, query server.QueryBundle) ([]output.Entry, error) {
	serverAddress := config.Hostname
	client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		return nil, err
	}

	// Synchronous call
	var reply server.QueryResult
	err = client.Call("Server.Leap", query, &reply)
	if err != nil {
		return nil, err
	}
	
	return reply.Entries, nil
}

func Shutdown(config *base.Configuration) error {
	serverAddress := config.Hostname
	client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		return err
	}

	var reply string
	err = client.Call("Server.Shutdown", "", &reply)
	if err != nil {
		return err
	}
	return nil
}

var token int

// ReIndexAndTransfer uses cindex to index a remote server's code. Then it
// transfers the index files to the local machine for faster queries.
func ReIndexAndTransfer(config *base.GlobalConfiguration) error {
	localproject := config.Currentproject
	localpath := config.Projects[localproject].Indexpath
	remoteproject := config.Projects[localproject].Remoteproject
	remotepath := config.Projects[localproject].Remotepath
	serverAddress := config.Projects[localproject].Host

	// this client thinger is what I want in the implementatino of the 
	// TODO(rjk): BlockSourceRequester needs an implementation of
	// a leapserver.
	leapserver, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		return err
	}

	// Token indentifies which ReIndexAndTransfer
	// TODO(rjk): Think through and document the semantics
	// for the why of Token.
	token++

	// I could specify the file and pathing? In particular,
	// the server version could have a simpler configuration?
	// this might make config easier? a remote wouldn't need as
	// much config?
	args := server.IndexAndBuildChecksumIndexArgs{
		Token: token,
		RemoteProjectName: remoteproject,
		RemotePath: remotepath,
	}
	var reply server.RemoteCheckSumIndexData

// there are two kinds of errors from the remote: where it's a connection failure
// or where the remote has successfully communicated a problem. We want
// to tell the remote that a sequence of transfer commands have completed?
	err = leapserver.Call("Server.IndexAndBuildChecksumIndex", args, &reply)
	if err != nil {
		return fmt.Errorf("Can't get remote to index and transfer because: %v", err)
		// close? cleanup? retry here?
	}
	fileSize := reply.FileSize

//	tell the remote if we can that it should clean up
//	defer client.Call

	// Compute the size locally (from the remote size)
	blockCount := fileSize / server.BLOCK_SIZE
	if fileSize%server.BLOCK_SIZE != 0 {
		blockCount++
	}

	// Setup the description of the remote file.
	fs := &gosync.BasicSummary{
		ChecksumIndex:  reply.ReferenceFileIndex,
		ChecksumLookup: reply.ChecksumLookup,
		BlockCount:     uint(blockCount),
		BlockSize:      uint(server.BLOCK_SIZE),
		FileSize:       fileSize,
	}

	// Construct a resolver.
	resolver := blocksources.MakeFileSizedBlockResolver(
		uint64(fs.GetBlockSize()),
		fs.GetFileSize(),
	)
	
	// Construct a BlockSource implementation (the way that blocks are
	// fetched from the remote.)
	blocksource := blocksources.NewBlockSourceBase(
		MakeRpcRequester(leapserver, token),
		resolver,
			&filechecksum.HashVerifier{
				Hash:                md5.New(),
				BlockSize:           fs.GetBlockSize(),
				BlockChecksumGetter: fs,
			},
		32,
		16*MB,
	)

	inputfile, err := os.Open(localpath)
	if err != nil {
		// TODO(rjk): Use details from the config
		return fmt.Errorf("can't open localpath because: %v", err)
	}
	defer inputfile.Close()

	outputfilename := filepath.Join(localpath, "-temporary")
	outputfile, err := os.Create(outputfilename)
	if err != nil {
		// TODO(rjk): Also use details from the config.
		return fmt.Errorf("can't create temporary output %s because %v", outputfilename, err)
	}
	defer outputfile.Close()

	// Build a RSync type to control the copying of the file from the remote.
	rsyncjob := &gosync.RSync  {
		// The local file that I want to update.
		Input: inputfile,

		// The remote supply of blocks. I implement this to transport blocks from
		// the remote to here. I use a composition-extended version of
		// BlockSourceBase
		Source: blocksource,

		// This is where we write the file. There will be three files: Input is
		// the local file that we want to update. Source is the remote file
		// (proxy) that will provide the updates. Output is where we
		// write the updated file (parts of Input and blocks from Source.)
		Output: outputfile,

		// The Summary describes the remote file. We build it above
		// from work done on the remote.
		Summary: fs,

		// Stuff to close.
		// TODO(rjk): I'm not sure how to define this.
		OnClose: nil,
	}

	// Actually ship files.
	if err := rsyncjob.Patch(); err != nil {
  	  return fmt.Errorf("rsync.Patch failed because %v", err)
	}

	// TODO(rjk): I presume that this runs the cleanup code?
	// I'm not sure yet how cleanup is suppose to work.
//	if err := rsync.Close(); err != nil {
//		// these might not be fatal?
//	}

	// TODO(rjk): Consider error-checking the close operations.
	outputfile.Close()
	inputfile.Close()

	replace(localpath, outputfilename)
	return nil
}


