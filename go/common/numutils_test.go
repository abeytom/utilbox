package common

import (
	"testing"
)

func TestIndices(t *testing.T) {
	AssertIntArray(t, ParseRange("10").Indices, []int{10})
	AssertIntArray(t, ParseRange("-10").Indices, []int{-10})

	AssertIntArray(t, ParseRange("10,20").Indices, []int{10, 20})
	AssertIntArray(t, ParseRange("10,-20").Indices, []int{10, -20})
	AssertIntArray(t, ParseRange("-10,-20").Indices, []int{-10, -20})

	AssertIntArray(t, CalcBounds("10-12"), []int{10, 12})
	AssertIntArray(t, CalcBounds("-10-12"), []int{-10, 12})
	AssertIntArray(t, CalcBounds("-10--12"), []int{-10, -12})
	AssertIntArray(t, CalcBounds("10--12"), []int{10, -12})

	AssertIntArray(t, ParseRange("10-12").Indices, []int{10, 11})
	AssertIntArray(t, ParseRange("-10-1").Indices, []int{-10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0})
	AssertIntArray(t, ParseRange("-10--12").Indices, []int{}) //invalid bounds
	AssertIntArray(t, ParseRange("-12--10").Indices, []int{-12, -11})
}

func TestStartEnd(t *testing.T) {
	assertRangeEquals(t, ParseRange("10:20"), P(10), P(20), nil)
	assertRangeEquals(t, ParseRange("10:"), P(10), nil, nil)
	assertRangeEquals(t, ParseRange(":20"), nil, P(20), nil)
}

func P(int2 int) *int {
	return &int2
}

func assertRangeEquals(t *testing.T, r *IntRange, start *int, end *int, indices []int) {
	if start != r.Start && *start != *r.Start {
		t.Fatalf("Start Mismatch Actual=%+v, Expected: %+v", start, *r.Start)
	}
	if end != r.End && *end != *r.End {
		t.Fatalf("End Mismatch Actual=%+v, Expected: %+v", start, *r.End)
	}
	AssertIntArray(t, r.Indices, indices)
}

func AssertIntArray(t *testing.T, actual []int, expected []int) {
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
