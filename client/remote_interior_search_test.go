package client

import (
	"encoding/xml"
	"net/rpc"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rjkroege/leap/base"
	"github.com/rjkroege/leap/output"
	"github.com/rjkroege/leap/search"
	"github.com/sanity-io/litter"
)

type querytest struct {
	name      string
	files     []string
	separator string
	suffix    []string
	result    []output.Entry
}

func leapIntegrationTestCore(t *testing.T, leapserver *rpc.Client, itd *IntegrationTestDirectory, allquerytests []querytest) {

	ris := &RemoteInternalSearcher{
		prefixes:    []string{itd.root},
		remoteindex: itd.remoteindexfile,
		leapserver:  leapserver,
	}
	searcher := search.NewTrigramSearch(itd.localindexfile, []string{itd.root})

	for _, tv := range allquerytests {
		queryresult, err := searcher.Query(tv.files, tv.separator, tv.suffix, ris)
		if err != nil {
			t.Errorf("Query %s failed: %v", tv.name, err)
		}
		if got, want := queryresult, tv.result; !reflect.DeepEqual(got, want) {
			t.Errorf("Query %s results got %s, want %s", tv.name, litter.Sdump(got), litter.Sdump(want))
		}
	}
}

func leapIntegrationTestBeforeReindex(t *testing.T, leapserver *rpc.Client, itd *IntegrationTestDirectory) {

	leapIntegrationTestCore(t, leapserver, itd, []querytest{
		{
			name:      "file name search for one",
			files:     []string{"one"},
			separator: ":",
			suffix:    []string{""},
			result: []output.Entry{
				{
					XMLName: xml.Name{
						Space: "",
						Local: "",
					},
					Uid:          filepath.Join(itd.root, "remote/one"),
					Arg:          filepath.Join(itd.root, "remote/one"),
					Type:         "file",
					Valid:        "",
					AutoComplete: "",
					Title:        "one",
					SubTitle:     "remote/one",
					Icon: output.AlfredIcon{
						Filename: "",
						Type:     "",
					},
				},
			},
		},

		{
			name:      "file name search inside",
			files:     []string{""},
			separator: ":/",
			suffix:    []string{"Hõla"},
			result: []output.Entry{
				{
					XMLName: xml.Name{
						Space: "",
						Local: "",
					},
					// Match is on line 1.
					Uid: filepath.Join(itd.root, "remote/two:1"),
					// TODO(rjk): There should not be a prepended /
					Arg:          "/" + filepath.Join(base.Prefix+":1", itd.root, "remote/two"),
					Type:         "file",
					Valid:        "",
					AutoComplete: "",
					Title:        "1 file two. Hõla",
					SubTitle:     ".../two:1 file two. Hõla",
					Icon: output.AlfredIcon{
						Filename: "",
						Type:     "",
					},
				},
			},
		},

		{
			name:      "file name search for (missing) four",
			files:     []string{"four"},
			separator: ":",
			suffix:    []string{""},
			result:    []output.Entry{},
		},

		// TODO(rjk): We are not consistent on missing matches. I should
		// be consistent.
		{
			name:      "file name search for (missing) contents of four",
			files:     []string{""},
			separator: ":/",
			suffix:    []string{"completely"},
			result:    nil,
		},
	})
}

func leapIntegrationTestAfterReindex(t *testing.T, leapserver *rpc.Client, itd *IntegrationTestDirectory) {

	leapIntegrationTestCore(t, leapserver, itd, []querytest{
		{
			name:      "file name search for one",
			files:     []string{"one"},
			separator: ":",
			suffix:    []string{""},
			result: []output.Entry{
				{
					XMLName: xml.Name{
						Space: "",
						Local: "",
					},
					Uid:          filepath.Join(itd.root, "remote/one"),
					Arg:          filepath.Join(itd.root, "remote/one"),
					Type:         "file",
					Valid:        "",
					AutoComplete: "",
					Title:        "one",
					SubTitle:     "remote/one",
					Icon: output.AlfredIcon{
						Filename: "",
						Type:     "",
					},
				},
			},
		},

		{
			name:      "file name search inside",
			files:     []string{""},
			separator: ":/",
			suffix:    []string{"Hõla"},
			result: []output.Entry{
				{
					XMLName: xml.Name{
						Space: "",
						Local: "",
					},
					// Match is on line 1.
					Uid: filepath.Join(itd.root, "remote/two:1"),
					// TODO(rjk): There should not be a prepended /
					Arg:          "/" + filepath.Join(base.Prefix+":1", itd.root, "remote/two"),
					Type:         "file",
					Valid:        "",
					AutoComplete: "",
					Title:        "1 file two. Hõla",
					SubTitle:     ".../two:1 file two. Hõla",
					Icon: output.AlfredIcon{
						Filename: "",
						Type:     "",
					},
				},
			},
		},

		{
			name:      "file name search for previously missing four",
			files:     []string{"four"},
			separator: ":",
			suffix:    []string{""},
			result: []output.Entry{
				{
					XMLName: xml.Name{
						Space: "",
						Local: "",
					},
					Uid:          filepath.Join(itd.root, "remote/newfourfile"),
					Arg:          filepath.Join(itd.root, "remote/newfourfile"),
					Type:         "file",
					Valid:        "",
					AutoComplete: "",
					Title:        "newfourfile",
					SubTitle:     "remote/newfourfile",
					Icon: output.AlfredIcon{
						Filename: "",
						Type:     "",
					},
				},
			},
		},

		// TODO(rjk): We are not consistent on missing matches. I should
		// be consistent.
		{
			name:      "content search satisfied by previously missing four",
			files:     []string{""},
			separator: ":/",
			suffix:    []string{"completely"},
			result: []output.Entry{
				{
					XMLName: xml.Name{
						Space: "",
						Local: "",
					},
					Uid:          filepath.Join(itd.root, "remote/newfourfile:2"),
					Arg:          "/" + filepath.Join(base.Prefix+":2", itd.root, "remote/newfourfile"),
					Type:         "file",
					Valid:        "",
					AutoComplete: "",
					Title:        "2 for something completely different\n",
					SubTitle:     ".../newfourfile:2 for something completely different\n",
					Icon: output.AlfredIcon{
						Filename: "",
						Type:     "",
					},
				},
			},
		},
	})
}
