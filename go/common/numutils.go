package common

import (
	"strconv"
	"strings"
	"unicode"
)

type IntRange struct {
	Start   *int
	End     *int
	Indices []int
	Exclude bool
}

/*
	[1]
	[1,2]
	[1,-3] => invalid but will parse
	[1-3]
	[1-3,10--12]
	[-1-3] == (-1) -> (+3)
	[-10--3] == (-10) -> (-3)
	[1,2,5:,6:]
*/
func ParseRange(str string) *IntRange {
	num, err := strconv.Atoi(str)
	if err == nil {
		return &IntRange{Indices: []int{num}}
	}
	var vals []int
	if strings.Index(str, ":") != -1 {
		split := strings.Split(str, ":")
		if len(split) == 1 {
			return &IntRange{Start: StrToIntP(split[0], nil)}
		} else {
			return &IntRange{Start: StrToIntP(split[0], nil), End: StrToIntP(split[1], nil)}
		}
	} else if strings.Index(str, ",") != -1 {
		parts := strings.Split(str, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			vals = append(vals, ParseRange(part).Indices...)
		}
		return &IntRange{Indices: vals}
	} else {
		atoi, err := strconv.Atoi(str)
		if err == nil {
			return &IntRange{Indices: []int{atoi}}
		}
		bounds := CalcBounds(str)
		return &IntRange{Indices: ResolveBounds(bounds)}
	}
}

func ResolveBounds(bounds []int) []int {
	if len(bounds) == 1 {
		return bounds
	}
	if len(bounds) != 2 {
		return []int{}
	}
	lower := bounds[0]
	upper := bounds[1]
	if upper <= lower {
		return []int{}
	}
	var vals []int
	for i := lower; i < upper; i++ {
		vals = append(vals, i)
	}
	return vals
}

func CalcBounds(str string) []int {
	runes := []rune(strings.TrimSpace(str))
	neg := false
	var chars []rune
	var bounds []int
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		if char == '-' {
			if i == 0 {
				neg = true
			} else {
				first, err := strconv.Atoi(string(chars))
				if err != nil {
					return []int{}
				}
				if neg {
					first = first * -1
				}
				bounds = append(bounds, first)
				//
				chars = nil
				nextChar := runes[i+1]
				if nextChar == '-' {
					neg = true
					i++
				} else if unicode.IsDigit(nextChar) {
					neg = false
				}
				//this could be a range-separator or negative-sign
			}
		} else if unicode.IsDigit(char) {
			chars = append(chars, char)
		}
	}
	if chars != nil {
		first, err := strconv.Atoi(string(chars))
		if err != nil {
			return []int{}
		}
		if neg {
			first = first * -1
		}
		bounds = append(bounds, first)
	}
	return bounds
}

func ResolveArrayBounds(r *IntRange, len int) []int {
	var start int
	if r.Start == nil {
		start = 0
	} else {
		start = *r.Start
	}

	var end int
	if r.End == nil {
		end = len
	} else {
		end = *r.End
	}
	var indices []int
	for i := start; i < end; i++ {
		indices = append(indices, i)
	}
	return indices
}
