package utils

import (
	"encoding/csv"
	"fmt"
	"github.com/abeytom/utilbox/common"
	"os"
	"strconv"
	"strings"
)

type LineProcessor struct {
	DataHeaders []string
	csvFmt      *CsvFormat
	RowIndex    int
	Lines       [][]string
	csvWriter   *CsvWriter
}

func NewLineProcessor(csvFmt *CsvFormat) *LineProcessor {
	processor := LineProcessor{csvFmt: csvFmt}
	if !csvFmt.HasWholeOpr {
		processor.csvWriter = NewCsvWriter(csvFmt)
	}
	return &processor
}

func (p *LineProcessor) processRow(supplier func() []string) {
	csvFmt := p.csvFmt
	if isWithInBounds(csvFmt.RowExt, p.RowIndex) {
		if p.RowIndex == 0 {
			if csvFmt.NoHeaderIn {
				//we consider this as a line
				words := extractCsv(supplier(), csvFmt.ColExt, csvFmt.ColFmtMap)
				if csvFmt.HasWholeOpr {
					p.Lines = append(p.Lines, words)
				} else {
					if !csvFmt.NoHeaderOut {
						p.csvWriter.WriteRaw(getFinalHeaders(csvFmt, nil))
					}
					p.csvWriter.Write(words)
				}
			} else {
				//this is a header
				words := extractCsv(supplier(), csvFmt.ColExt, nil)
				if csvFmt.HasWholeOpr {
					p.DataHeaders = words
				} else if !csvFmt.NoHeaderOut {
					p.csvWriter.WriteRaw(getFinalHeaders(csvFmt, words))
				}
			}
		} else {
			words := extractCsv(supplier(), csvFmt.ColExt, csvFmt.ColFmtMap)
			if csvFmt.HasWholeOpr {
				p.Lines = append(p.Lines, words)
			} else {
				p.csvWriter.Write(words)
			}
		}
	} else if p.RowIndex == 0 {
		if !csvFmt.NoHeaderIn {
			p.DataHeaders = extractCsv(supplier(), csvFmt.ColExt, nil)
		}
		if !csvFmt.HasWholeOpr && !csvFmt.NoHeaderOut {
			p.csvWriter.WriteRaw(getFinalHeaders(csvFmt, p.DataHeaders))
		}
	}
	p.RowIndex = p.RowIndex + 1
}

func (p *LineProcessor) Close() {
	if p.csvWriter != nil {
		p.csvWriter.Close()
	}
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
		return ConvertForMapping(nVal)
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
	case common.StringCol:
		if nVal != "" {
			oVal.(common.StringCol).Add(nVal)
		}
		return oVal
	default:
		//fixme this is an error; hanlde
		return ConvertForMapping(nVal)
	}
}

func Reduce(accumulated interface{}, nVal interface{}, strSep string) interface{} {
	if accumulated == nil {
		return nVal
	}
	//fmt.Println(">>")
	//fmt.Println(accumulated, strSep, nVal)
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

func ConvertForMapping(val string) interface{} {
	int64Val, err := strconv.ParseInt(val, 10, 64)
	if err == nil {
		return int64Val
	}
	float64Val, err := strconv.ParseFloat(val, 64)
	if err == nil {
		return float64Val
	}
	set := &common.StringSet{}
	if val != "" {
		set.Add(val)
	}
	return set
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
	return val
}

func ConvertIfNeeded(val interface{}) interface{} {
	if val == nil {
		return 0
	}
	switch val.(type) {
	case int64:
		return val
	case float64:
		return val
	case string:
		return Convert(val.(string))
	default:
		return val
	}
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
	showCount := groupMap.CsvFormat.MapRed.ShowCount
	for k, v := range groupMap.Map {
		keys := strings.Split(k, ":==:")
		colCount := keyCount + valueCount
		if showCount {
			colCount++
		}
		cols := make([]interface{}, colCount)
		for i, key := range keys {
			cols[i] = key
		}
		for i, val := range v.Values {
			cols[i+keyCount] = val
		}
		if showCount {
			cols[colCount-1] = int64(v.Count)
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
	for _, index := range s.Indices {
		compare := s.Compare(one[index], two[index])
		if compare == 0 {
			continue
		}
		if compare < 0 {
			return !s.Desc
		}
		return s.Desc
	}
	return true
}

func (s *DataRowSort) Compare(one interface{}, two interface{}) int {
	switch one.(type) {
	case int:
		twoVal := ConvInt(two, -1)
		if one.(int) == twoVal {
			return 0
		} else if one.(int) < twoVal {
			return -1
		} else {
			return 1
		}
	case int64:
		twoVal := ConvInt64(two, -1)
		if one.(int64) == twoVal {
			return 0
		} else if one.(int64) < twoVal {
			return -1
		} else {
			return 1
		}
	case float64:
		twoVal := ConvFloat64(two, -1)
		if one.(float64) == twoVal {
			return 0
		} else if one.(float64) < twoVal {
			return -1
		} else {
			return 1
		}
	default:
		str1 := fmt.Sprintf("%v", one)
		str2 := fmt.Sprintf("%v", two)
		if str1 == str2 {
			return 0
		} else if str1 < str2 {
			return -1
		} else {
			return 1
		}
	}
}

func ConvInt(val interface{}, def int) int {
	switch val.(type) {
	case int:
		return val.(int)
	case int64:
		return int(val.(int64))
	case float64:
		return int(val.(float64))
	default:
		str := fmt.Sprintf("%v", val)
		atoi, err := strconv.Atoi(str)
		if err == nil {
			return atoi
		}
		return def
	}
}

func ConvInt64(val interface{}, def int64) int64 {
	switch val.(type) {
	case int:
		return int64(val.(int))
	case int64:
		return val.(int64)
	case float64:
		return int64(val.(float64))
	default:
		str := fmt.Sprintf("%v", val)
		atoi, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			return atoi
		}
		return def
	}
}

func ConvFloat64(val interface{}, def float64) float64 {
	switch val.(type) {
	case int:
		return float64(val.(int))
	case int64:
		return float64(val.(int64))
	case float64:
		return val.(float64)
	default:
		str := fmt.Sprintf("%v", val)
		atoi, err := strconv.ParseFloat(str, 64)
		if err == nil {
			return atoi
		}
		return def
	}
}

type CsvWriter struct {
	CsvFormat *CsvFormat
	CsvWriter *csv.Writer
}

func NewCsvWriter(format *CsvFormat) *CsvWriter {
	writer := CsvWriter{CsvFormat: format}
	if format.Merge == "csv" || (format.OutputDef != nil && format.OutputDef.Type == "csv") {
		writer.CsvWriter = csv.NewWriter(os.Stdout)
	}
	return &writer
}

func (w *CsvWriter) WriteAll(dataRows []DataRow) {
	for _, row := range dataRows {
		w.WriteRow(&row)
	}
	w.Close()
}

func (w *CsvWriter) Write(words []string) {
	if w.CsvFormat.CalcDefs != nil {
		cols := make([]interface{}, len(words))
		for i, v := range words {
			cols[i] = v
		}
		nRow := applyCalc(w.CsvFormat, &DataRow{Cols: cols})
		w.WriteRow(nRow)
	} else {
		w.WriteRaw(words)
	}
}

func (w *CsvWriter) WriteRow(row *DataRow) {
	var vals []string
	for _, word := range row.Cols {
		str := common.ToString(word)
		vals = append(vals, str)
	}
	w.WriteRaw(vals)
}

func (w *CsvWriter) WriteRaw(words []string) {
	if w.CsvWriter != nil {
		err := w.CsvWriter.Write(words)
		if err != nil {
			//todo handle error
		}
	} else {
		csvFormat := w.CsvFormat
		line := strings.Join(words, csvFormat.Merge)
		if csvFormat.Wrap != "" {
			line = csvFormat.Wrap + line + csvFormat.Wrap
		}
		fmt.Printf("%s\n", line)
	}
}

func (w *CsvWriter) Close() {
	if w.CsvWriter != nil {
		w.CsvWriter.Flush()
	}
}
