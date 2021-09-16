package utils

import "strings"

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

func ProcessGroupByOutput() {

}

type GroupMapValue struct {
	Values []interface{}
	Count  int
}

func (mapVal *GroupMapValue) Append(values []string) {
	if mapVal.Values == nil {
		mapVal.Values = make([]interface{}, len(values))
	}
	mapVal.Count = mapVal.Count + 1
}

type GroupMap struct {
	Map map[string]GroupMapValue
}

func (groupMap *GroupMap) put(keys []string, values []string) {
	key := strings.Join(keys, ":==:")
	mapVal, exists := groupMap.Map[key]
	if !exists {
		mapVal = GroupMapValue{}
		groupMap.Map[key] = mapVal
	}
	mapVal.Append(values)
}
