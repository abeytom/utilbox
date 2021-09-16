package utils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/abeytom/utilbox/common"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ColumnFormat struct {
	Delim  string
	Merge  string
	ColExt *common.IntRange
	Wrap   string
	Prefix string
	Suffix string
}

type CsvInjectArgs struct {
	Prefix string
	Suffix string
}

type CsvFormat struct {
	ColFmtMap    map[int]ColumnFormat
	ColExt       *common.IntRange
	RowExt       *common.IntRange
	Delim        string
	Merge        string
	IsLMerge     bool
	LMerge       string
	Wrap         string
	HasReducer   bool
	MapRed       *MapRed
	HasMrReducer bool
}

func HandleCsv(args []string) {
	//csv [delimiter]  [merge] [row_def] [col_def]

	//delimiter -> space(default) tab comma
	//merge -> default false
	//row_def -> row[0:] (default), row[1] row[2:3]
	//col_def -> col[0:] (default), col[1] col[2:3]

	// kc get pods | csv space row[1:] col[0]
	// kc get pods | csv space  col[0] -> gets all rows
	// kc get pods | csv space -> all lines will be merged
	// kc get pods | csv space merge ->
	//fmt.Printf("The args are %s\n", args)

	filePath := args[0]
	csvFmt := CsvFormat{
		ColFmtMap: map[int]ColumnFormat{},
		ColExt:    &common.IntRange{},
		RowExt:    &common.IntRange{},
		Delim:     " ",
		Merge:     ",",
		LMerge:    ",",
		Wrap:      "",
		IsLMerge:  false,
	}
	//kc get svc | csv row[1:] col[0-3,5] fmt#c4#split:/#merge:-#col[0,1] fmt#c2#split:.#merge:: merge:'|' lmerge:===
	//kc get svc | csv row[1:] col[0,4] merge:: fmt.c0.pfx:'curl http://'  fmt.c4.split:/.col[0].sfx:'/actuator/health' wrap:dquote
	// kc get pods | csv row[1:] col[0] fmt#c0#split:-#:-#ncol[-1,-2]
	// kc get svc | csv row[1:] | csv split:csv merge:'|' lmerge:'     >>>>>     '
	//todo filter row
	//todo sum , sort, group
	//todo replace-chars

	//cat ~/tmp/topics.txt | csv row[1:] col[2,3,4,6] mr#group:[3]
	//cat ~/tmp/topics.txt | csv row[1:] col[6,2,3,4] mr#group:[0]#sort:[1]
	//cat ~/tmp/topics.txt | csv row[1:] col[6,2,3,4] mr#group:[0]#sort:[1]:desc
	//cat ~/tmp/topics.txt | csv row[1:] col[6,2,3,4] mr#group:[0]#sort:[0]:asc#sum:row
	//cat ~/tmp/topics.txt | csv row[1:] col[2,3,4] mr#sum:row
	//cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr#group:[0,1]#sort:[1]:asc merge:tab => 2 group by keys
	//cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr#group:[0,1]#sort:[2]:desc#sum:row

	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "lmerge") {
			csvFmt.IsLMerge = true
			if arg != "lmerge" {
				csvFmt.LMerge = extractDelim(arg, "lmerge:")
			}
		} else if strings.HasPrefix(arg, "merge") {
			csvFmt.Merge = extractDelim(arg, "merge:")
		} else if strings.Index(arg, "row[") == 0 {
			csvFmt.RowExt = extractCsvIndexArg(arg)
		} else if strings.Index(arg, "col[") == 0 {
			csvFmt.ColExt = extractCsvIndexArg(arg)
		} else if strings.Index(arg, "ncol[") == 0 {
			csvFmt.ColExt = extractCsvIndexArg(arg)
			csvFmt.ColExt.Exclude = true
		} else if strings.Index(arg, "split:") == 0 {
			csvFmt.Delim = extractDelim(arg, "split:")
		} else if strings.Index(arg, "wrap:") == 0 {
			csvFmt.Wrap = extractDelim(arg, "wrap:")
		} else if strings.HasPrefix(arg, "fmt") {
			processFmtArguments(arg, &csvFmt)
		} else if strings.HasPrefix(arg, "mr") {
			processMrArguments(arg, &csvFmt)
		}
	}
	csvFmt.HasReducer = hasReducer(csvFmt)
	csvFmt.HasMrReducer = hasMrReducer(csvFmt.MapRed)

	//fmt.Printf("[%+v] \n", csvFmt)

	file, err := os.Open(filePath)
	if err != nil {
		panic("Cannot read the file " + filePath)
	}
	defer file.Close()

	if csvFmt.Delim == "csv" {
		processCsv(file, csvFmt)
	} else {
		scanner := bufio.NewScanner(file)
		processor := LineProcessor{csvFmt: &csvFmt}
		for scanner.Scan() {
			processor.processRow(func() []string {
				return strings.Split(scanner.Text(), csvFmt.Delim)
			})
		}
		processLines(&csvFmt, &processor)
	}
}

func processCsv(file *os.File, csvFmt CsvFormat) {
	reader := csv.NewReader(bufio.NewReader(file))
	processor := LineProcessor{csvFmt: &csvFmt}
	for {
		words, _ := reader.Read()
		if words == nil {
			break
		}
		processor.processRow(func() []string { return words })
	}
	processLines(&csvFmt, &processor)
}

func processLines(csvFmt *CsvFormat, processor *LineProcessor) {
	if len(processor.Lines) <= 0 {
		return
	}
	if csvFmt.IsLMerge {
		var lines []string
		for _, words := range processor.Lines {
			lines = append(lines, strings.Join(words, csvFmt.Merge))
		}
		fmt.Printf("%s\n", strings.Join(lines, csvFmt.LMerge))
	} else if csvFmt.MapRed != nil {
		if csvFmt.MapRed.GroupBy != nil {
			applyGroupBy(csvFmt, processor, csvFmt.MapRed)
		} else if csvFmt.MapRed.Sum == "row" {
			applyRowSum(csvFmt, processor.Lines)
		} else {
			applyColSum(csvFmt, processor)
		}
	}
}

func applyRowSum(csvFmt *CsvFormat, lines [][]string) {
	for _, words := range lines {
		var rowSum int64
		for _, word := range words {
			val, err := strconv.ParseInt(word, 10, 64)
			if err != nil {
				//fmt.Println("Cannot convert %s to number ", err)
			} else {
				rowSum += val
			}
		}
		printLine([]string{strconv.FormatInt(rowSum, 10)}, csvFmt)
	}
}

func applyColSum(csvFmt *CsvFormat, processor *LineProcessor) {
	cols := len(processor.Lines[0])
	var row = make([]int64, cols)
	for _, words := range processor.Lines {
		for i, word := range words {
			val, err := strconv.ParseInt(word, 10, 64)
			if err != nil {
				//fmt.Println("Cannot convert %s to number ", err)
			} else {
				row[i] = row[i] + val
			}
		}
	}
	var words = make([]string, cols)
	for i, val := range row {
		words[i] = strconv.FormatInt(val, 10)
	}
	printLine(words, csvFmt)
}

func pickWords(words []string, indices []int) []string {
	var vals []string
	for _, index := range indices {
		vals = append(vals, words[index])
	}
	return vals
}

func applyGroupBy(csvFmt *CsvFormat, processor *LineProcessor, mapRed *MapRed) {
	firstRow := processor.Lines[0]
	//compute the groupByIndices
	groupBy := mapRed.GroupBy
	groupByIndices := common.GetFilterStrIndices(common.ApplyRange(firstRow, groupBy))

	//compute non-groupByIndices
	var nonGroupByIndices []int
	for i := 0; i < len(firstRow); i++ {
		if !common.BruteIntContains(groupByIndices, i) {
			nonGroupByIndices = append(nonGroupByIndices, i)
		}
	}
	groupMap := make(map[string][]int64)
	for _, words := range processor.Lines {
		groupByKey := strings.Join(pickWords(words, groupByIndices), ":==:")
		row, ok := groupMap[groupByKey]
		if !ok {
			row = make([]int64, len(firstRow)-len(groupByIndices))
			groupMap[groupByKey] = row
		}
		valWords := pickWords(words, nonGroupByIndices)
		for i, word := range valWords {
			val, err := strconv.ParseInt(word, 10, 64)
			if err != nil {
				//fmt.Println("Cannot convert %s to number ", err)
			} else {
				row[i] = row[i] + val
			}
		}
	}
	hasReducer := hasMrReducer(mapRed)
	var array [][]interface{}
	for k, v := range groupMap {
		keys := strings.Split(k, ":==:")
		keyCount := len(keys)
		if mapRed.Sum == "row" {
			var rowSum int64
			for _, val := range v {
				rowSum += val
			}
			if hasReducer {
				words := make([]interface{}, keyCount+1)
				for i, key := range keys {
					words[i] = key
				}
				words[keyCount] = rowSum
				array = append(array, words)
			} else {
				words := make([]string, keyCount+1)
				for i, key := range keys {
					words[i] = key
				}
				words[keyCount] = strconv.FormatInt(rowSum, 10)
				printLine(words, csvFmt)
			}
		} else {
			if hasReducer {
				words := make([]interface{}, keyCount+len(v))
				for i, key := range keys {
					words[i] = key
				}
				for i, val := range v {
					words[i+keyCount] = val
				}
				array = append(array, words)
			} else {
				words := make([]string, keyCount+len(v))
				for i, key := range keys {
					words[i] = key
				}
				for i, val := range v {
					words[i+keyCount] = strconv.FormatInt(val, 10)
				}
				printLine(words, csvFmt)
			}
		}
	}
	if hasReducer {
		applySort(array, csvFmt)
	}
}

type ISort struct {
	Items   [][]interface{}
	Indices []int
	Desc    bool
}

func (s *ISort) Len() int {
	return len(s.Items)
}
func (s *ISort) Swap(i, j int) {
	items := s.Items
	items[i], items[j] = items[j], items[i]
}
func (s *ISort) Less(i, j int) bool {
	one := s.Items[i]
	two := s.Items[j]
	index := s.Indices[0] //todo multi sort cols
	compare := s.Compare(one[index], two[index])
	if s.Desc {
		return !compare
	}
	return compare
}

func (s *ISort) Compare(one interface{}, two interface{}) bool {
	switch one.(type) {
	case int64:
		return one.(int64) < two.(int64)
	default:
		return fmt.Sprintf("%v", one) < fmt.Sprintf("%v", two)
	}
}

func applySort(lines [][]interface{}, csvFmt *CsvFormat) {
	sortDef := csvFmt.MapRed.SortDef
	sortCols := sortDef.SortCols
	applyRange := *common.IApplyRange(lines[0], sortCols)
	var indices []int
	for _, item := range applyRange {
		indices = append(indices, item.Index)
	}
	iSort := &ISort{
		Items:   lines,
		Indices: indices,
		Desc:    sortDef.Desc,
	}
	sort.Sort(iSort)
	for _, line := range lines {
		var vals []string
		for _, word := range line {
			vals = append(vals, fmt.Sprintf("%v", word))
		}
		printLine(vals, csvFmt)
	}
}

func printLine(words []string, csvFmt *CsvFormat) {
	line := strings.Join(words, csvFmt.Merge)
	if csvFmt.Wrap != "" {
		line = csvFmt.Wrap + line + csvFmt.Wrap
	}
	fmt.Printf("%s\n", line)
}

func hasReducer(csvFmt CsvFormat) bool {
	if csvFmt.IsLMerge || csvFmt.MapRed != nil {
		return true
	}
	return false
}

//func processWords(words []string, csvFmt *CsvFormat) []string {
//	var vals []string
//	for _, word := range words {
//		vals = append(vals, word)
//	}
//	//line := strings.Join(vals, csvFmt.Merge)
//	//if csvFmt.Wrap != "" {
//	//	return csvFmt.Wrap + line + csvFmt.Wrap
//	//}
//	return vals
//}

func hasMrReducer(mapRed *MapRed) bool {
	return mapRed != nil && mapRed.SortDef != nil
}

type MapRed struct {
	Sum     string
	GroupBy *common.IntRange
	SortDef *SortDef
}

type SortDef struct {
	SortCols *common.IntRange
	Desc     bool
}

func processMrArguments(command string, csvFmt *CsvFormat) {
	chars := []rune(command)
	sep := string(chars[2])
	parts := strings.Split(command, sep)
	mapRed := MapRed{}
	for _, part := range parts[1:] {
		if strings.Index(part, "sum") != -1 {
			split := strings.Split(part, ":")
			if len(split) == 1 {
				mapRed.Sum = "col"
			} else {
				mapRed.Sum = split[1]
			}
		} else if strings.Index(part, "group:") != -1 {
			split := strings.Split(part, ":")[1:]
			mapRed.GroupBy = extractCsvIndexArg(split[0])
		} else if strings.Index(part, "sort:") != -1 {
			split := strings.Split(part, ":")[1:]
			sort := &SortDef{}
			sort.SortCols = extractCsvIndexArg(split[0])
			if len(split) > 1 {
				if split[1] == "desc" {
					sort.Desc = true
				} else {
					sort.Desc = false
				}
			}
			mapRed.SortDef = sort
		}
	}
	csvFmt.MapRed = &mapRed
}

func processFmtArguments(command string, csvFmt *CsvFormat) {
	chars := []rune(command)
	sep := string(chars[3])
	parts := strings.Split(command, sep)
	colIndex, err := strconv.Atoi(strings.Replace(parts[1], "c", "", 1))
	if err != nil {
		panic("Invalid col index for formatting" + parts[1])
	}
	format := ColumnFormat{}
	for _, part := range parts[2:] {
		if strings.Index(part, "split:") == 0 {
			format.Delim = extractDelim(part, "split:")
		} else if strings.Index(part, "merge:") == 0 {
			format.Merge = extractDelim(part, "merge:")
		} else if strings.Index(part, "wrap:") == 0 {
			format.Wrap = extractDelim(part, "wrap:")
		} else if strings.Index(part, "sfx:") == 0 {
			format.Suffix = extractDelim(part, "sfx:")
		} else if strings.Index(part, "pfx:") == 0 {
			format.Prefix = extractDelim(part, "pfx:")
		} else if strings.Index(part, "col[") == 0 {
			format.ColExt = extractCsvIndexArg(part)
		} else if strings.Index(part, "ncol[") == 0 {
			format.ColExt = extractCsvIndexArg(part)
			format.ColExt.Exclude = true
		}
	}
	csvFmt.ColFmtMap[colIndex] = format
}

func extractDelim(arg string, prefix string) string {
	delim := strings.Replace(arg, prefix, "", 1)
	if delim == "comma" {
		return ","
	}
	if delim == "space" {
		return " "
	}
	if delim == "tab" {
		return "\t"
	}
	if delim == "newline" {
		return "\n"
	}
	if delim == "quote" {
		return "'"
	}
	if delim == "dquote" {
		return "\""
	}
	if delim == "none" {
		return ""
	}
	if delim == "csv" {
		return "csv"
	}
	if delim == "pipe" {
		return "pipe"
	}
	return delim
}

func extractCsv(words []string, ext *common.IntRange, fmtMap map[int]ColumnFormat) []string {
	ftrWords := common.ApplyRange(words, ext)
	var vals []string
	for _, ftrWord := range *ftrWords {
		word := processWord(ftrWord, fmtMap)
		vals = append(vals, word)
	}
	return vals
}

func processWord(ftrWord common.FilterStr, fmtMap map[int]ColumnFormat) string {
	word := ftrWord.Str
	colFormat, ok := fmtMap[ftrWord.Index]
	if !ok {
		return word
	}
	if colFormat.Delim != "" {
		parts := strings.Split(word, colFormat.Delim)
		if colFormat.ColExt != nil {
			parts = extractCsv(parts, colFormat.ColExt, map[int]ColumnFormat{})
		}
		merge := ","
		if colFormat.Merge != "" {
			merge = colFormat.Merge
		}
		word = strings.Join(parts, merge)
	}
	if colFormat.Prefix != "" {
		word = colFormat.Prefix + word
	}
	if colFormat.Suffix != "" {
		word = word + colFormat.Suffix
	}
	if colFormat.Wrap != "" {
		word = colFormat.Wrap + word + colFormat.Wrap
	}
	return word
}

func isWithInBounds(ext *common.IntRange, idx int) bool {
	upper := true
	lower := true
	if ext.Start != nil {
		upper = idx >= *ext.Start
	}
	if ext.End != nil {
		if ext.Start != ext.End {
			lower = idx < *ext.End
		} else {
			lower = idx == *ext.End
		}
	}
	return upper && lower
}

func extractCsvIndexArg(arg string) *common.IntRange {
	start := strings.Index(arg, "[")
	end := strings.Index(arg, "]")
	if start >= 0 && end > start {
		str := string(([]rune(arg))[start+1 : end])
		return common.ParseRange(str)
	} else {
		return &common.IntRange{}
	}
}
