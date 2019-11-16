package index

import (
	"log"
	"fmt"

	"github.com/codeskyblue/go-sh"
	"github.com/rjkroege/leap/base"
)

type Idx struct{}

// ReIndex runs cindex to reindex the file based on the provided configuration.
// cindex has to be in the path.
// TODO(rjk): Validate the args from the client.
// TODO(rjk): Assume less config state? It's not clear where the args should
// come from here. 
func (_ Idx) ReIndex(config *base.GlobalConfiguration, currentproject string) ([]byte, error) {
	log.Printf("Invoked reindex %v %v\n" , config, currentproject)
	log.Println("currentproject is", currentproject)

	session := sh.NewSession()
	indexpath := config.Projects[currentproject].Indexpath
	log.Println("indexpath: ", indexpath)
	session.SetEnv("CSEARCHINDEX", indexpath)

	output, err := session.Command("cindex").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("can't run cindex because: %v", err)
	}
	return output, nil
}
