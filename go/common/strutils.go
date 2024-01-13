package common

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
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
			var val string
			if len(array) > index {
				val = array[index]
			} else {
				val = ""
			}
			vals = append(vals, FilterStr{Str: val, Index: index})
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
	for _, item := range *filters {
		indices = append(indices, item.Index)
	}
	return indices
}

func GetFilterItemIndices(filters *[]FilterItem) []int {
	var indices []int
	for _, item := range *filters {
		indices = append(indices, item.Index)
	}
	return indices
}

func ParseSubCommandArg(arg string) []string {
	chars := []rune(arg)
	in := 0
	prev := 0
	var args []string
	for i, char := range chars {
		if char == ':' {
			if in == 0 {
				r := chars[prev:i]
				args = append(args, string(r))
				prev = i + 1
			}
		} else if char == '[' {
			in++
		} else if char == ']' {
			in--
		}
	}
	length := len(chars)
	if prev < length-1 {
		args = append(args, string(chars[prev:length]))
	}
	return args
}

func ParseIndexStr(arg string) []string {
	start := strings.Index(arg, "[")
	end := strings.Index(arg, "]")
	if start >= 0 && end > start {
		str := string(([]rune(arg))[start+1 : end])
		return strings.Split(str, ",")
	}
	return nil
}

func ParseExprStr(arg string) string {
	start := strings.Index(arg, "(")
	end := strings.LastIndex(arg, ")")
	if start >= 0 && end > start {
		return string(([]rune(arg))[start+1 : end])
	}
	return ""
}

func ToString(word interface{}) string {
	var str string
	switch word.(type) {
	case StringCol:
		str = (word.(StringCol)).ToString()
	case float64:
		str = fmt.Sprintf("%.2f", word)
		if strings.Index(str, ".00") == len(str)-3 { //hack
			str = string([]rune(str)[0 : len(str)-3])
		}
	case int64:
		str = fmt.Sprintf("%v", word)
	default:
		str = fmt.Sprintf("%v", word)
	}
	return str
}

var void struct{}

type StringCol interface {
	Add(str string)
	Values() []string
	ToString() string
	MarshalJSON() ([]byte, error)
}

type StringList struct {
	values []string
}

func NewStringList() *StringList {
	return &StringList{}
}

func (s *StringList) Add(str string) {
	s.values = append(s.values, str)
}
func (s *StringList) Values() []string {
	return s.values
}

func (s *StringList) ToString() string {
	return strings.Join(s.Values(), ",")
}

func (s *StringList) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

// todo avoid saving keys if it is not needed
type StringSet struct {
	values  map[string]struct{}
	keys    []string
	ordered bool
}

func NewStringSet(vals []string) *StringSet {
	values := make(map[string]struct{})
	for _, val := range vals {
		values[val] = void

	}
	return &StringSet{values: values, keys: vals}
}

func NewOrderedStringSet(vals []string) *StringSet {
	set := NewStringSet(vals)
	set.ordered = true
	return set
}

func (s *StringSet) Add(str string) {
	if s.values == nil {
		s.values = make(map[string]struct{})
	}
	if _, exists := s.values[str]; !exists {
		s.values[str] = void
		s.keys = append(s.keys, str)
	}
}

func (s *StringSet) Values() []string {
	if s.ordered {
		return s.keys
	}
	keys := make([]string, len(s.values))
	i := 0
	for k := range s.values {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func (s *StringSet) ToString() string {
	return strings.Join(s.Values(), ",")
}

func (s *StringSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values())
}

func DelimToCamelCase(str string, delim rune, capitalizeFirst bool) string {
	isToUpper := false
	var output string
	for k, v := range str {
		if k == 0 {
			if capitalizeFirst {
				output = strings.ToUpper(string(v))
			} else {
				output = string(v)
			}
		} else {
			if isToUpper {
				output += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == delim {
					isToUpper = true
				} else {
					output += string(v)
				}
			}
		}
	}
	return output
}

func DelBlankItems(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
