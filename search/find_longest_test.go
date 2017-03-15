package search

import (
	"testing"
)

func TestFindLongestPrefixOfOne(t *testing.T) {
	in := []string{"hi"}

	if expected, got := 2, findLongestPrefix(in); expected != got {
		t.Errorf("got %v expected %v\n", got, expected)
	}

}

func TestFindLongestPrefixsSingleChar(t *testing.T) {
	in := []string{"hi", "h"}

	if expected, got := 1, findLongestPrefix(in); expected != got {
		t.Errorf("got %v expected %v\n", got, expected)
	}

}

func TestFindLongestPrefixsDoubleChar(t *testing.T) {
	in := []string{"hi", "he"}

	if expected, got := 1, findLongestPrefix(in); expected != got {
		t.Errorf("got %v expected %v\n", got, expected)
	}

}

func TestFindLongestPrefixsNoChar(t *testing.T) {
	in := []string{"hi", "om"}

	if expected, got := 0, findLongestPrefix(in); expected != got {
		t.Errorf("got %v expected %v\n", got, expected)
	}

}

func TestFindLongmatch(t *testing.T) {
	in := []string{"hit", "hit"}

	if expected, got := 3, findLongestPrefix(in); expected != got {
		t.Errorf("got %v expected %v\n", got, expected)
	}

}
