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

	allQuery := &index.Query{Op: index.QAll}
	post := ix.PostingQuery(allQuery)
	
//	fnames := make([]uint32, 0, len(post))

	for _, fileid := range post {
		name := ix.Name(fileid)
		log.Printf("name %#v", name)
//		if fre.MatchString(name, true, true) < 0 {
//			continue
//		}
//		fnames = append(fnames, fileid)
	}
	return nil, nil
}