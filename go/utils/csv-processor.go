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
	Header    []string
	csvFmt    *CsvFormat
	RowIndex  int
	Lines     [][]string
	csvWriter *CsvWriter
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
		words := extractCsv(supplier(), csvFmt.ColExt, csvFmt.ColFmtMap)
		if p.RowIndex == 0 {
			p.Header = words
			if !csvFmt.HasWholeOpr {
				PrintHeader(p, p.csvWriter)
			}
		}
		if csvFmt.HasWholeOpr {
			p.Lines = append(p.Lines, words)
		} else {
			p.csvWriter.Write(words)
		}
	} else if p.RowIndex == 0 {
		p.Header = extractCsv(supplier(), csvFmt.ColExt, csvFmt.ColFmtMap)
		if !csvFmt.HasWholeOpr {
			PrintHeader(p, p.csvWriter)
		}
	}
	p.RowIndex = p.RowIndex + 1
}

func (p *LineProcessor) Close() {
	if p.csvWriter != nil {
		p.csvWriter.Close()
	}
}

func GetHeaders(p *LineProcessor) []string {
	if p.csvFmt.HeaderDef != nil && p.csvFmt.HeaderDef.Fields != nil {
		return p.csvFmt.HeaderDef.Fields
	} else {
		return p.Header
	}
}

func PrintHeader(p *LineProcessor, writer *CsvWriter) {
	if p.csvFmt.HeaderDef != nil {
		headers := GetHeaders(p)
		if headers != nil {
			writer.WriteRaw(applyCalcHeaders(writer.CsvFormat, headers))
		}
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
	case *common.StringSet:
		oVal.(*common.StringSet).Add(nVal)
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
	set.Add(val)
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
