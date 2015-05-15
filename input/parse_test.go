package input

import (
	"testing"
)

func TestParse(t *testing.T) {

	a, b := Parse("ab")
	if !(a == "a" && b == "b") {
		t.Errorf("got %v,%v, exepcted %v, %v", a, b, "a", "b")
	}
}
