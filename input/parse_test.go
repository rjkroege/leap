package input

import (
	"testing"
)

func TestChunkInput(t *testing.T) {

	a, s, b := chunkInput("ab")
	if ea, es, eb := "ab", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %v,%v, exepcted %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("ab/c")
	if ea, es, eb := "ab", "/", "c"; a != ea || b != eb || s != es {
		t.Errorf("got %v,%v, exepcted %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("ab#c")
	if ea, es, eb := "ab", "#", "c"; a != ea || b != eb || s != es {
		t.Errorf("got %v,%v, exepcted %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("ab@c")
	if ea, es, eb := "ab", "@", "c"; a != ea || b != eb || s != es {
		t.Errorf("got %v,%v, exepcted %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("@c")
	if ea, es, eb := "", "@", "c"; a != ea || b != eb || s != es {
		t.Errorf("got %v,%v, exepcted %v, %v", a, s, b, ea, es, eb)
	}

}

func TestFileExp(t *testing.T) {
	
	if a, ea  := fileExp("a"), ".*a.*" ; a != ea  {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	

	if a, ea  := fileExp("ab"), ".*a.*b.*" ; a != ea  {
		t.Errorf("got %v exepcted %v", a, ea)
	}

}

func TestNumCheck(t *testing.T) {
	if a, ea  := numCheck("a"), "" ; a != ea  {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	

	if a, ea  := numCheck("23"), "23" ; a != ea  {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	
}