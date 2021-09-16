package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type LineProcessor struct {
	csvFmt   *CsvFormat
	RowIndex int
	Lines    [][]string
}

func (p *LineProcessor) processRow(supplier func() []string) {
	csvFmt := p.csvFmt
	if isWithInBounds(csvFmt.RowExt, p.RowIndex) {
		words := extractCsv(supplier(), csvFmt.ColExt, csvFmt.ColFmtMap)
		//words = processWords(words, csvFmt)
		if csvFmt.HasReducer {
			p.Lines = append(p.Lines, words)
		} else {
			printLine(words, csvFmt)
		}
	}
	p.RowIndex = p.RowIndex + 1
}

type OutputProcessor struct {
	CsvFormat *CsvFormat
}

type GroupMapValue struct {
	Values []interface{}
	Count  int
}

func (mapVal *GroupMapValue) Append(values []string) {
	if mapVal.Values == nil {
		mapVal.Values = make([]interface{}, len(values))
	}
	for i, value := range values {
		mapVal.Values[i] = Merge(mapVal.Values[i], value)
	}
	mapVal.Count = mapVal.Count + 1
}

func Merge(oVal interface{}, nVal string) interface{} {
	if oVal == nil {
		return Convert(nVal)
	}
	switch oVal.(type) {
	case int64:
		int64Val, err := strconv.ParseInt(nVal, 10, 64)
		if err == nil {
			return int64Val + oVal.(int64)
		}
		return oVal
	case float64:
		float64Val, err := strconv.ParseFloat(nVal, 64)
		if err == nil {
			return float64Val + oVal.(float64)
		}
		return oVal
	case []string:
		return append(oVal.([]string), nVal)
	default:
		//fixme this is an error; hanlde
		return Convert(nVal)
	}
}

func Reduce(accumulated interface{}, nVal interface{}, strSep string) interface{} {
	if accumulated == nil {
		return nVal
	}
	switch accumulated.(type) {
	case int64:
		return accumulated.(int64) + nVal.(int64)
	case float64:
		return accumulated.(float64) + nVal.(float64)
	case string:
		return accumulated.(string) + strSep + nVal.(string)
	default:
		return accumulated
	}
}

func Convert(val string) interface{} {
	int64Val, err := strconv.ParseInt(val, 10, 64)
	if err == nil {
		return int64Val
	}
	float64Val, err := strconv.ParseFloat(val, 64)
	if err == nil {
		return float64Val
	}
	return []string{val}
}

type GroupMap struct {
	Map          map[string]*GroupMapValue
	KeyIndices   []int
	ValueIndices []int
	CsvFormat    *CsvFormat
}

func (groupMap *GroupMap) Put(keys []string, values []string) {
	if groupMap.Map == nil {
		groupMap.Map = make(map[string]*GroupMapValue)
	}
	key := strings.Join(keys, ":==:")
	mapVal, exists := groupMap.Map[key]
	if !exists {
		mapVal = &GroupMapValue{}
		groupMap.Map[key] = mapVal
	}
	mapVal.Append(values)
}

func (groupMap *GroupMap) PostProcess() []DataRow {
	if groupMap.CsvFormat.MapRed == nil || groupMap.CsvFormat.MapRed.Sum != "row" {
		return groupMap.Flatten()
	}
	newMap := make(map[string]*GroupMapValue)
	for k, value := range groupMap.Map {
		var merged interface{}
		for i, val := range value.Values {
			value.Values[i] = Reduce(merged, val, groupMap.CsvFormat.Merge)
		}
		newMap[k] = value
	}
	groupMap.Map = newMap
	return groupMap.Flatten()
}

type DataRow struct {
	Cols  []interface{}
	Count int
}

func (groupMap *GroupMap) Flatten() []DataRow {
	keyCount := len(groupMap.KeyIndices)
	valueCount := len(groupMap.ValueIndices)
	var array []DataRow
	for k, v := range groupMap.Map {
		keys := strings.Split(k, ":==:")
		cols := make([]interface{}, keyCount+valueCount)
		for i, key := range keys {
			cols[i] = key
		}
		for i, val := range v.Values {
			cols[i+keyCount] = val
		}
		array = append(array, DataRow{
			Cols:  cols,
			Count: v.Count,
		})
	}
	return array
}

type DataRowSort struct {
	Rows    []DataRow
	Indices []int
	Desc    bool
}

func (s *DataRowSort) Len() int {
	return len(s.Rows)
}
func (s *DataRowSort) Swap(i, j int) {
	items := s.Rows
	items[i], items[j] = items[j], items[i]
}
func (s *DataRowSort) Less(i, j int) bool {
	one := s.Rows[i].Cols
	two := s.Rows[j].Cols
	index := s.Indices[0] //todo multi sort cols
	compare := s.Compare(one[index], two[index])
	if s.Desc {
		return !compare
	}
	return compare
}

func (s *DataRowSort) Compare(one interface{}, two interface{}) bool {
	switch one.(type) {
	case int64:
		return one.(int64) < two.(int64)
	case float64:
		return one.(float64) < two.(float64)
	default:
		return fmt.Sprintf("%v", one) < fmt.Sprintf("%v", two)
	}
}
