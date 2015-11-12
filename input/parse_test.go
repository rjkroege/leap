package input

import (
	"reflect"
	"regexp"
	"testing"
)

func TestChunkInput(t *testing.T) {
	a, s, b := chunkInput("ab")
	if ea, es, eb := "ab", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v",
			a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("a")
	if ea, es, eb := "a", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v",
			a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("a/b/c")
	if ea, es, eb := "a/b/c", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v",
			a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("/a/b/c")
	if ea, es, eb := "/a/b/c", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v",
			a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("/a/b/c/:/de")
	if ea, es, eb := "/a/b/c/", ":", "/de"; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v",
			a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("a/b")
	if ea, es, eb := "a/b", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v",
			a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("a/b:/c")
	if ea, es, eb := "a/b", ":", "/c"; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput(":/c")
	if ea, es, eb := "", ":", "/c"; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("a/b#c")
	if ea, es, eb := "a/b", "#", "c"; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("a/b@c")
	if ea, es, eb := "a/b", "@", "c"; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("@c")
	if ea, es, eb := "", "@", "c"; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("")
	if ea, es, eb := "", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("da/")
	if ea, es, eb := "da/", "", ""; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}

	a, s, b = chunkInput("da/:/cd")
	if ea, es, eb := "da/", ":", "/cd"; a != ea || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v exepcted %#v, %#v %#v", a, s, b, ea, es, eb)
	}
}

func TestFileExp(t *testing.T) {

	if a, ea := fileExp("a"), ".*a.*"; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	if _, err := regexp.Compile(fileExp("a")); err != nil {
		t.Errorf("invalid regexp: %v", err)
	}

	if a, ea := fileExp("ab"), ".*a.*b.*"; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	if _, err := regexp.Compile(fileExp("ab")); err != nil {
		t.Errorf("invalid regexp: %v", err)
	}

	if a, ea := fileExp("a/b"), ".*a/b.*"; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	if _, err := regexp.Compile(fileExp("a/b")); err != nil {
		t.Errorf("invalid regexp: %v", err)
	}

	if a, ea := fileExp("a/bc/d"), ".*a/b.*c/d.*"; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	if _, err := regexp.Compile(fileExp("a/bc/d")); err != nil {
		t.Errorf("invalid regexp: %v", err)
	}
}

func TestNumCheck(t *testing.T) {
	if a, ea := numCheck("a"), ""; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}

	if a, ea := numCheck("23"), "23"; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}

}

func TestSymbolExp(t *testing.T) {
	if a, ea := symbolExp("a"), "(func|type|var|const).*a[a-zA-Z_0-9]*"; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	if _, err := regexp.Compile(symbolExp("a")); err != nil {
		t.Errorf("invalid regexp: %v", err)
	}

	if a, ea := symbolExp("ab"), "(func|type|var|const).*a[a-zA-Z_0-9]*b[a-zA-Z_0-9]*"; a != ea {
		t.Errorf("got %v exepcted %v", a, ea)
	}
	if _, err := regexp.Compile(symbolExp("ab")); err != nil {
		t.Errorf("invalid regexp: %v", err)
	}
}

func TestParse(t *testing.T) {
	a, s, b := Parse("a:/b")
	if ea, es, eb := []string{"a[^/]*$", "a", "^a", "a", ".*a.*"}, "/", ".*b.*"; !reflect.DeepEqual(a, ea) || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v, exepcted %v, %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = Parse("a@b")
	if ea, es, eb := []string{"a[^/]*$", "a", "^a", "a", ".*a.*"}, "/", "(func|type|var|const).*b[a-zA-Z_0-9]*";  !reflect.DeepEqual(a, ea) || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v, exepcted %v, %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = Parse("a:10")
	if ea, es, eb := []string{"a[^/]*$", "a", "^a", "a", ".*a.*"}, ":", "10";  !reflect.DeepEqual(a, ea) || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v, exepcted %v, %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = Parse("a")
	if ea, es, eb := []string{"a[^/]*$", "a", "^a", "a", ".*a.*"}, ":", "";  !reflect.DeepEqual(a, ea) || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v, exepcted %v, %v, %v", a, s, b, ea, es, eb)
	}

	a, s, b = Parse("a/")
	if ea, es, eb := []string{"a/[^/]*$", "a/", "^a[^/]*/", "a[^/]*/", ".*a/.*"}, ":", "";  !reflect.DeepEqual(a, ea) || b != eb || s != es {
		t.Errorf("got %#v,%#v, %#v, exepcted %v, %v, %v", a, s, b, ea, es, eb)
	}
}

func TestFuzzyMatchers(t *testing.T) {
	a := fuzzyMatchers("abc")
	if ea := []string{"abc[^/]*$", "abc", "^abc", "abc", ".*a.*b.*c.*"}; !reflect.DeepEqual(a, ea) {
		t.Errorf("got %#v, exepcted %#v", a,ea)
	}

	a = fuzzyMatchers("abc/def")
	if ea := []string{"abc/def[^/]*$", "abc/def", "^abc[^/]*/def", "abc[^/]*/def", ".*a.*b.*c/d.*e.*f.*"}; !reflect.DeepEqual(a, ea) {
		t.Errorf("got %#v, exepcted %#v", a,ea)
	}

	a = fuzzyMatchers("abc/")
	if ea := []string{"abc/[^/]*$", "abc/", "^abc[^/]*/", "abc[^/]*/", ".*a.*b.*c/.*"}; !reflect.DeepEqual(a, ea) {
		t.Errorf("got %#v, exepcted %#v", a,ea)
	}
}
