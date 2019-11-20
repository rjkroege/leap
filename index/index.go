package index

import (
	"fmt"
	"log"

	"github.com/codeskyblue/go-sh"
)

type Idx struct{}

// ReIndex runs cindex to reindex the file based on the provided configuration.
// cindex has to be in the path.
// TODO(rjk): Validate the args from the client.
// TODO(rjk): Assume less config state? It's not clear where the args should
// come from here.
func (_ Idx) ReIndex(indexpath string, args ...string) ([]byte, error) {
	session := sh.NewSession()
	log.Println("indexpath: ", indexpath)
	session.SetEnv("CSEARCHINDEX", indexpath)

	// TODO(rjk): insert argument here.
	output, err := session.Command("cindex", args).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("can't run cindex because: %v", err)
	}
	return output, nil
}
