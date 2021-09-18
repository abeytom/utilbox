package common

import "testing"

func TestParseSubArgs(t *testing.T) {
	AssertStrArray(t, ParseSubCommandArg("col"), []string{"col"})
	AssertStrArray(t, ParseSubCommandArg("col[10:20]"), []string{"col[10:20]"})
	AssertStrArray(t, ParseSubCommandArg("col[10:20]:"), []string{"col[10:20]"})
	AssertStrArray(t, ParseSubCommandArg("col[10:20]:arg2"), []string{"col[10:20]", "arg2"})
	AssertStrArray(t, ParseSubCommandArg("col[10:20]:arg2:group[2:3]:group[3:]"), []string{"col[10:20]", "arg2", "group[2:3]", "group[3:]"})
}

func AssertStrArray(t *testing.T, actual []string, expected []string) {
	if len(actual) != len(expected) {
		t.Fatalf("Array Mismatch Actual=%+v, Expected: %+v", actual, expected)
	}
	for index, _ := range actual {
		if actual[index] != expected[index] {
			t.Fatalf("Array Mismatch at Index %d Actual=%+v, Expected: %+v",
				index, actual[index], expected[index])
		}
	}
}
