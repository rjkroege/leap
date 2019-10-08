package index

import (
	"log"

	"github.com/codeskyblue/go-sh"
	"github.com/rjkroege/leap/base"
)

// ReIndex runs cindex to reindex the
func ReIndex(config *base.GlobalConfiguration) {
	log.Println("Invoked reindex")

	session := sh.NewSession()
	indexpath := config.Projects[config.Currentproject].Indexpath
	log.Println("indexpath: ", indexpath)
	session.SetEnv("CSEARCHINDEX", indexpath)
	if err := session.Command("cindex").Run(); err != nil {
		log.Println("Couldn't re-index: ", err)
	}
}
