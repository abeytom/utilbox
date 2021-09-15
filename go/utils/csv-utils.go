package utils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/abeytom/utilbox/common"
	"os"
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
	ColFmtMap  map[int]ColumnFormat
	ColExt     *common.IntRange
	RowExt     *common.IntRange
	Delim      string
	Merge      string
	IsLMerge   bool
	LMerge     string
	Wrap       string
	HasReducer bool
}

type LineProcessor struct {
	csvFmt   CsvFormat
	RowIndex int
	Lines    []string
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
			processFmtArguments(arg, csvFmt)
		}
	}
	csvFmt.HasReducer = hasReducer(csvFmt)

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
		processor := LineProcessor{csvFmt: csvFmt}
		for scanner.Scan() {
			processor.processRow(scanner.Text())
		}
		if csvFmt.IsLMerge {
			fmt.Printf("%s\n", strings.Join(processor.Lines, csvFmt.LMerge))
		}
	}
}

func processCsv(file *os.File, csvFmt CsvFormat) {
	reader := csv.NewReader(bufio.NewReader(file))
	processor := LineProcessor{csvFmt: csvFmt}
	for {
		words, _ := reader.Read()
		if words == nil {
			break
		}
		processor.processRowCols(words)
	}

	if csvFmt.IsLMerge {
		fmt.Printf("%s\n", strings.Join(processor.Lines, csvFmt.LMerge))
	}
}

func (p *LineProcessor) processRow(line string) {
	csvFmt := p.csvFmt
	if isWithInBounds(csvFmt.RowExt, p.RowIndex) {
		words := strings.Split(line, csvFmt.Delim)
		words = extractCsv(words, csvFmt.ColExt, csvFmt.ColFmtMap)
		pLine := processWords(words, csvFmt)
		if csvFmt.HasReducer {
			p.Lines = append(p.Lines, pLine)
		} else {
			fmt.Printf("%s\n", pLine)
		}
	}
	p.RowIndex = p.RowIndex + 1
}

func (p *LineProcessor) processRowCols(words []string) {
	csvFmt := p.csvFmt
	if isWithInBounds(csvFmt.RowExt, p.RowIndex) {
		words = extractCsv(words, csvFmt.ColExt, csvFmt.ColFmtMap)
		line := processWords(words, csvFmt)
		if csvFmt.HasReducer {
			p.Lines = append(p.Lines, line)
		} else {
			fmt.Printf("%s\n", line)
		}
	}
	p.RowIndex = p.RowIndex + 1
}

func hasReducer(csvFmt CsvFormat) bool {
	if csvFmt.IsLMerge {
		return true
	}
	return false
}

func processWords(words []string, csvFmt CsvFormat) string {
	var vals []string
	for _, word := range words {
		vals = append(vals, word)
	}
	line := strings.Join(vals, csvFmt.Merge)
	if csvFmt.Wrap != "" {
		return csvFmt.Wrap + line + csvFmt.Wrap
	}
	return line
}

func processFmtArguments(command string, csvFmt CsvFormat) {
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
	if start > 0 && end > start {
		str := string(([]rune(arg))[start+1 : end])
		return common.ParseRange(str)
	} else {
		return &common.IntRange{}
	}
}
