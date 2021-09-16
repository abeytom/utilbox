package common

import (
	"strconv"
)

type FilterStr struct {
	Str   string
	Index int
}

type FilterItem struct {
	Item  interface{}
	Index int
}

func ApplyRange(array []string, intR *IntRange) *[]FilterStr {
	var indices []int
	if intR.Indices == nil {
		indices = ResolveArrayBounds(intR, len(array))
	} else {
		indices = intR.Indices
	}
	if !intR.Exclude {
		var vals []FilterStr
		for _, index := range indices {
			if index < 0 {
				index = len(array) + index
			}
			vals = append(vals, FilterStr{Str: array[index], Index: index})
		}
		return &vals
	} else {
		idxSet := make(map[int]bool)
		for _, index := range indices {
			if index < 0 {
				index = len(array) + index
			}
			idxSet[index] = true
		}
		var vals []FilterStr
		for index, item := range array {
			_, ok := idxSet[index]
			if !ok { //not contains
				vals = append(vals, FilterStr{Str: item, Index: index})
			}
		}
		return &vals
	}
}

func IApplyRange(array []interface{}, intR *IntRange) *[]FilterItem {
	var indices []int
	if intR.Indices == nil {
		indices = ResolveArrayBounds(intR, len(array))
	} else {
		indices = intR.Indices
	}
	if !intR.Exclude {
		var vals []FilterItem
		for _, index := range indices {
			if index < 0 {
				index = len(array) + index
			}
			vals = append(vals, FilterItem{Item: array[index], Index: index})
		}
		return &vals
	} else {
		idxSet := make(map[int]bool)
		for _, index := range indices {
			if index < 0 {
				index = len(array) + index
			}
			idxSet[index] = true
		}
		var vals []FilterItem
		for index, item := range array {
			_, ok := idxSet[index]
			if !ok { //not contains
				vals = append(vals, FilterItem{Item: item, Index: index})
			}
		}
		return &vals
	}
}

func StrToInt(str string, def int) int {
	parsed, err := strconv.Atoi(str)
	if err != nil {
		return def
	} else {
		return parsed
	}
}

func StrToIntP(str string, def *int) *int {
	parsed, err := strconv.Atoi(str)
	if err != nil {
		return def
	} else {
		return &parsed
	}
}

func GetFilterStrIndices(filters *[]FilterStr) []int {
	var indices []int
	for i, _ := range *filters {
		indices = append(indices, i)
	}
	return indices
}
