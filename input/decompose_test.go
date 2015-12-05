package input

import (
	"testing"

	"github.com/rjkroege/leap/base"
)

func TestEncodedToPlumb(t *testing.T) {
	if a, ea :=  EncodedToPlumb("/ab"), "/ab";  a != ea {
		t.Errorf("got %#v exepcted %#v", a, ea)
	}
	
	if a, ea :=  EncodedToPlumb("/" + base.Prefix + ":100/ab"), "/ab:100";  a != ea {
		t.Errorf("got %#v exepcted %#v", a, ea)
	}

	if a, ea :=  EncodedToPlumb("/tmp/.leaping/glenda:5/Users/rjkroege/tools/gopkg/src/github.com/rjkroege/leap/main.go"), "/Users/rjkroege/tools/gopkg/src/github.com/rjkroege/leap/main.go:5";  a != ea {
		t.Errorf("got %#v exepcted %#v", a, ea)
	}
}

func TestEncodedToFile(t *testing.T) {
	if a, ea :=  EncodedToFile("/ab"), "/ab";  a != ea {
		t.Errorf("got %#v exepcted %#v", a, ea)
	}
	
	if a, ea :=  EncodedToFile("/" + base.Prefix + ":100/ab"), "/ab";  a != ea {
		t.Errorf("got %#v exepcted %#v", a, ea)
	}
}
