package search

import (
	"os"
	"log"

	"github.com/google/codesearch/index"
	"github.com/rjkroege/leap/output"
)

type filenameSearch struct {
	index.Index
}

func NewFileNameSearch() output.Generator {
	return &filenameSearch{ *index.Open(index.File()) }
}


func (ix *filenameSearch) innerQuery(i int) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("innerQuery failed at %d", i)
			os.Exit(-1)
		}
	}()
	return ix.Name(uint32(i))
}
 
func (ix *filenameSearch) Query(fn, qtype, suffix string) ([]output.Entry, error) {
	//	compile the regexp
	//re, err := regexp.Compile(fn)
	//if err != nil {
	//	return nil, err
	//}

	// oo := make(output.Entry,0)

	// use the query-all : var allQuery = &Query{Op: QAll}
	// then, iterate...

	// OK. 
	//       for each file name, stuff result into array.
	for i := 0;  true ; i++ {
		name := ix.innerQuery(i)
		log.Printf("id %d %#v", i, name)
		if name == "\000" || name == "" {
			break
		}
	}
	return nil, nil
}