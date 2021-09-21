package utils

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/abeytom/utilbox/common"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ColumnFormat struct {
	Split  string
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
	ColFmtMap    map[int][]ColumnFormat
	ColExt       *common.IntRange
	RowExt       *common.IntRange
	Split        string
	Merge        string
	IsLMerge     bool
	LMerge       string
	Wrap         string
	HasWholeOpr  bool
	MapRed       *GroupByDef
	HasMrReducer bool
	OutputDef    *OutputDef
	SortDef      *SortDef
	HeaderDef    *HeaderDef
	CalcDefs     []CalcDef
}

type GroupByDef struct {
	Sum        string //deprecated
	ColIndices *common.IntRange
	ShowCount  bool
	//SortDef *SortDef
}

type SortDef struct {
	SortCols *common.IntRange
	Desc     bool
}

type OutputDef struct {
	Type   string
	Fields []string
	Levels int
}

type HeaderDef struct {
	Fields []string
}

type CalcDef struct {
	RawExpr    string
	ParsedExpr string
	Indices    *common.IntSet
	EvalExpr   *govaluate.EvaluableExpression
	FieldName  string
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
	csvFmt := &CsvFormat{
		ColExt:   &common.IntRange{},
		RowExt:   &common.IntRange{},
		Split:    " ",
		Merge:    "csv",
		LMerge:   ",",
		Wrap:     "",
		IsLMerge: false,
	}
	//kc get svc | csv row[1:] col[0-3,5] fmt#c4#split:/#merge:-#col[0,1] fmt#c2#split:.#merge:: merge:'|' lmerge:===
	//kc get svc | csv row[1:] col[0,4] merge:: fmt.c0.pfx:'curl http://'  fmt.c4.split:/.col[0].sfx:'/actuator/health' wrap:dquote
	// kc get pods | csv row[1:] col[0] fmt#c0#split:-#:-#ncol[-1,-2]
	// kc get svc | csv row[1:] | csv split:csv merge:'|' lmerge:'     >>>>>     '
	//todo filter row
	//todo sum , sort, group
	//todo replace-chars

	//cat ~/tmp/topics.txt | csv row[1:] col[2,3,4,6] mr#group[3]
	//cat ~/tmp/topics.txt | csv row[1:] col[6,2,3,4] mr#group[0]#sort[1]
	//cat ~/tmp/topics.txt | csv row[1:] col[6,2,3,4] mr#group[0]#sort[1]:desc
	//cat ~/tmp/topics.txt | csv row[1:] col[6,2,3,4] mr#group[0]#sort[0]:asc#sum:row
	//cat ~/tmp/topics.txt | csv row[1:] col[2,3,4] mr#sum:row
	//cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr#group[0,1]#sort[1]:asc merge:tab => 2 group by keys
	//cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr#group[0,1]#sort[2]:desc#sum:row
	// cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr,,group[0,1],,sort[2] out..json..cols:topic:groups:group:in:out:lag
	//cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr,,group[0,1],,sort[2] out..json..cols:topic:groups:group:in:out:lag => One Group
	//cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr,,group[0,1],,sort[2] out..json..cols:topic:groups:group:stats:in:out:lag => 2 Level JSON
	//cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr,,group[0:2],,sort[2]:desc out..json..fields[topic,groups,group,in,out,lag]
	/*
		# Chained format on the same column
		cat ~/kdev_default_svc.txt | csv  out#csv fmt#c4#split:comma#col[-1]  fmt#c4#split:/#col[0] fmt#c4#split::#col[-1]
		# merged string fields
		cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr..group[0]..sort[2]:desc out..json
		cat ~/tmp/topics.txt | csv row[1:] col[0,6,2,3,4] mr..group[0]..sort[2]:desc out..json..levels:1
	*/

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
		} else if strings.Index(arg, "sort[") == 0 {
			csvFmt.SortDef = extractSort(arg)
		} else if strings.Index(arg, "split:") == 0 {
			csvFmt.Split = extractDelim(arg, "split:")
		} else if strings.Index(arg, "wrap:") == 0 {
			csvFmt.Wrap = extractDelim(arg, "wrap:")
		} else if strings.HasPrefix(arg, "fmt") {
			processFmtArguments(arg, csvFmt)
		} else if strings.HasPrefix(arg, "group[") {
			processGroupArgs(arg, csvFmt)
		} else if strings.HasPrefix(arg, "out") {
			processOutputArgs(arg, csvFmt)
		} else if strings.HasPrefix(arg, "head") {
			csvFmt.HeaderDef = extractHeaderDef(arg)
		} else if strings.HasPrefix(arg, "calc") {
			extractCalcDef(arg, csvFmt)
		}
		//else if strings.HasPrefix(arg, "mr") {
		//	processMrArguments(arg, csvFmt)
		//}
	}
	csvFmt.HasWholeOpr = hasWholeOpr(csvFmt)
	csvFmt.HasMrReducer = hasMrReducer(csvFmt.MapRed)

	//fmt.Printf("[%+v] \n", csvFmt)

	file, err := os.Open(filePath)
	if err != nil {
		panic("Cannot read the file " + filePath)
	}
	defer file.Close()

	if csvFmt.Split == "csv" {
		processCsv(file, csvFmt)
	} else {
		scanner := bufio.NewScanner(file)
		processor := NewLineProcessor(csvFmt)
		for scanner.Scan() {
			processor.processRow(func() []string {
				return strings.Split(scanner.Text(), csvFmt.Split)
			})
		}
		processLines(csvFmt, processor)
		processor.Close()
	}
}

func processCsv(file *os.File, csvFmt *CsvFormat) {
	reader := csv.NewReader(bufio.NewReader(file))
	processor := NewLineProcessor(csvFmt)
	for {
		words, _ := reader.Read()
		if words == nil {
			break
		}
		processor.processRow(func() []string { return words })
	}
	processLines(csvFmt, processor)
	processor.Close()
}

func processLines(csvFmt *CsvFormat, processor *LineProcessor) {
	if len(processor.Lines) <= 0 {
		return
	}
	if csvFmt.MapRed != nil {
		if csvFmt.MapRed.ColIndices != nil {
			applyGroupBy(csvFmt, processor, csvFmt.MapRed)
		} else if csvFmt.MapRed.Sum == "row" {
			applyRowSum(csvFmt, processor)
		} else {
			applyColSum(csvFmt, processor)
		}
	} else {
		dataRows := toDataRows(processor)
		dataRows = applyCalcAll(csvFmt, dataRows)
		if csvFmt.SortDef != nil {
			dataRows = convertAndApplySort(csvFmt, dataRows)
		}
		if csvFmt.IsLMerge {
			//todo use csv to create row
			var lines []string
			for _, row := range dataRows {
				var words []string
				for _, word := range row.Cols {
					var str string
					switch word.(type) {
					case *common.StringSet:
						str = (word.(*common.StringSet)).ToString()
					default:
						str = fmt.Sprintf("%v", word)
					}
					words = append(words, str)
				}
				lines = append(lines, strings.Join(words, csvFmt.Merge))
			}
			fmt.Printf("%s\n", strings.Join(lines, csvFmt.LMerge))
		} else if csvFmt.OutputDef != nil {
			def := csvFmt.OutputDef
			if def.Type == "json" {
				csvFmt.OutputDef.Levels = 0
				headers := applyCalcHeaders(csvFmt, GetHeaders(processor))
				processJsonOutput(dataRows, csvFmt, headers, 0)
			} else if def.Type == "table" {
				headers := applyCalcHeaders(csvFmt, GetHeaders(processor))
				processTableOutput(dataRows, csvFmt, headers)
			} else {
				printLines(csvFmt, processor, dataRows)
			}
		} else {
			printLines(csvFmt, processor, dataRows)
		}
	}
}

func printLines(csvFmt *CsvFormat, processor *LineProcessor, dataRows []DataRow) {
	writer := NewCsvWriter(csvFmt)
	PrintHeader(processor, writer)
	writer.WriteAll(dataRows)
	writer.Close()
}

func applyRowSum(csvFmt *CsvFormat, processor *LineProcessor) {
	writer := NewCsvWriter(csvFmt)
	if csvFmt.HeaderDef != nil {
		headers := GetHeaders(processor)
		header := "(" + strings.Join(headers, " + ") + ")" //fixme todo apply headers
		writer.Write([]string{header})
	}

	for _, words := range processor.Lines {
		var rowSum int64
		for _, word := range words {
			val, err := strconv.ParseInt(word, 10, 64)
			if err != nil {
				//fmt.Println("Cannot convert %s to number ", err)
			} else {
				rowSum += val
			}
		}
		writer.Write([]string{strconv.FormatInt(rowSum, 10)})
	}
	writer.Close()
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
	writer := NewCsvWriter(csvFmt)
	var words = make([]string, cols)
	for i, val := range row {
		words[i] = strconv.FormatInt(val, 10)
	}
	PrintHeader(processor, writer)
	writer.Write(words)
	writer.Close()
}

func pickWords(words []string, indices []int) []string {
	var vals []string
	for _, index := range indices {
		vals = append(vals, words[index])
	}
	return vals
}

func applyGroupBy(csvFmt *CsvFormat, processor *LineProcessor, mapRed *GroupByDef) {
	firstRow := processor.Lines[0]
	//compute the keyIndices
	groupBy := mapRed.ColIndices
	keyIndices := common.GetFilterStrIndices(common.ApplyRange(firstRow, groupBy))
	//compute valueIndices
	var valueIndices []int
	for i := 0; i < len(firstRow); i++ {
		if !common.BruteIntContains(keyIndices, i) {
			valueIndices = append(valueIndices, i)
		}
	}
	groupMap := GroupMap{
		KeyIndices:   keyIndices,
		ValueIndices: valueIndices,
		CsvFormat:    csvFmt,
	}
	for _, words := range processor.Lines {
		keys := pickWords(words, keyIndices)
		values := pickWords(words, valueIndices)
		groupMap.Put(keys, values)
	}
	dataRows := applyCalcAll(csvFmt, groupMap.PostProcess())
	if csvFmt.SortDef != nil {
		applySort(dataRows, csvFmt)
	}
	headers := applyGroupByHeaders(csvFmt, processor.Header, keyIndices)
	headers = applyCalcHeaders(csvFmt, headers)
	if csvFmt.OutputDef != nil {
		if csvFmt.OutputDef.Type == "json" {
			processJsonOutput(dataRows, csvFmt, headers, len(keyIndices))
		} else if csvFmt.OutputDef.Type == "table" {
			processTableOutput(dataRows, csvFmt, headers)
		} else {
			printCsv(csvFmt, headers, dataRows)
		}
		return
	}
	printCsv(csvFmt, headers, dataRows)
}

func printCsv(csvFmt *CsvFormat, headers []string, dataRows []DataRow) {
	writer := NewCsvWriter(csvFmt)
	if csvFmt.HeaderDef != nil {
		writer.WriteRaw(headers)
	}
	writer.WriteAll(dataRows)
	writer.Close()
}

func processTableOutput(rows []DataRow, csvFmt *CsvFormat, headers []string) {
	colWidths := make(map[int]int)
	for _, row := range rows {
		for i, col := range row.Cols {
			var width int
			switch col.(type) {
			case *common.StringSet:
				for _, val := range col.(*common.StringSet).Values() {
					if (len(val)) > width {
						width = len(val)
					}
				}
			default:
				width = len(common.ToString(col))
			}
			existing := colWidths[i]
			if width > existing {
				colWidths[i] = width
			}
		}
	}
	for i, header := range headers {
		existing := colWidths[i]
		if len(header) > existing {
			colWidths[i] = len(header)
		}
	}
	fmtMap := make(map[int]string)
	for k, v := range colWidths {
		fmtMap[k] = "%-" + strconv.Itoa(v+3) + "s "
	}
	for i, header := range headers {
		fmt.Printf(fmtMap[i], header)
	}
	fmt.Println("")
	for _, row := range rows {
		var extraCols [][]string
		extraValCount := 0
		for i, col := range row.Cols {
			switch col.(type) {
			case *common.StringSet:
				vals := col.(*common.StringSet).Values()
				if len(vals) == 0 {
					fmt.Printf(fmtMap[i], "")
				} else if len(vals) == 1 {
					fmt.Printf(fmtMap[i], common.ToString(vals[0]))
				} else {
					fmt.Printf(fmtMap[i], common.ToString(vals[0]))

					extraCols = make([][]string, len(row.Cols))
					extraColVals := make([]string, len(vals)-1)
					for i := 1; i < len(vals); i++ {
						extraColVals[i-1] = vals[i]
					}
					extraCols[i] = extraColVals
					if extraValCount < len(extraColVals) {
						extraValCount = len(extraColVals)
					}
				}
			default:
				fmt.Printf(fmtMap[i], common.ToString(col))
			}
		}
		fmt.Println("")
		if len(extraCols) > 0 {
			for i := 0; i < extraValCount; i++ {
				for j, vals := range extraCols {
					if len(vals) <= i {
						fmt.Printf(fmtMap[j], "")
					} else {
						fmt.Printf(fmtMap[j], vals[i])
					}
				}
				fmt.Println("")
			}
		}
	}
}

func applyGroupByHeaders(csvFmt *CsvFormat, headers []string, keyIndices []int) []string {
	if csvFmt.HeaderDef != nil && csvFmt.HeaderDef.Fields != nil {
		return csvFmt.HeaderDef.Fields
	}

	var nHeaders []string
	added := make(map[int]bool)
	for _, index := range keyIndices {
		nHeaders = append(nHeaders, headers[index])
		added[index] = true
	}
	for i, h := range headers {
		if _, exists := added[i]; !exists {
			nHeaders = append(nHeaders, h)
		}
	}
	if csvFmt.MapRed.ShowCount {
		nHeaders = append(nHeaders, "count")
	}
	return nHeaders
}

func applyCalcHeaders(csvFmt *CsvFormat, headers []string) []string {
	calcDefs := csvFmt.CalcDefs
	if len(calcDefs) == 0 {
		return headers
	}
	//if headers are set by the user, then use that
	if csvFmt.HeaderDef != nil && csvFmt.HeaderDef.Fields != nil {
		return headers
	}
	var calcHeaders []string
	for _, calcDef := range calcDefs {
		indexSet := calcDef.Indices
		indexVals := indexSet.Values()
		headExpr := calcDef.ParsedExpr
		for _, index := range indexVals {
			headExpr = strings.Replace(headExpr, fmt.Sprintf("col%d", index), headers[index], 1)
		}
		calcHeaders = append(calcHeaders, headExpr)
	}
	return append(headers, calcHeaders...)

	//indexSet := calcDef.Indices
	//indexVals := indexSet.Values()
	//headExpr := calcDef.ParsedExpr
	//for _, index := range indexVals {
	//	headExpr = strings.Replace(headExpr, fmt.Sprintf("col%d", index), headers[index], 1)
	//}
	//var nHeaders []string
	//added := false
	//for i, col := range headers {
	//	if !indexSet.Contains(i) {
	//		nHeaders = append(nHeaders, col)
	//	} else {
	//		if !added {
	//			nHeaders = append(nHeaders, headExpr)
	//			added = true
	//		}
	//	}
	//}
	//return nHeaders
}

func applyCalcAll(csvFmt *CsvFormat, rows []DataRow) []DataRow {
	calcDef := csvFmt.CalcDefs
	if calcDef == nil {
		return rows
	}
	nRows := make([]DataRow, len(rows))
	for i, row := range rows {
		nRows[i] = *applyCalc(csvFmt, &row)
	}
	return nRows
}

func applyCalc(csvFmt *CsvFormat, row *DataRow) *DataRow {
	calcDefs := csvFmt.CalcDefs
	if len(calcDefs) == 0 {
		return row
	}
	var calcValues []interface{}
	for _, calcDef := range calcDefs {
		params := govaluate.MapParameters{}
		indexSet := calcDef.Indices
		for i, col := range row.Cols {
			if indexSet.Contains(i) {
				key := fmt.Sprintf("col%d", i)
				//todo if any args are string, then dont convert into number
				params[key] = ConvertIfNeeded(col)
				//fmt.Printf("%T:%v\n", params[key], params[key])
			}
		}

		eval, err := calcDef.EvalExpr.Eval(params)
		if err != nil {
			eval = err.Error()
		}
		switch eval.(type) {
		case float64:
			if !hasFloatArgs(params) {

			}
		}
		//fmt.Printf("EVAL: %T:%v\n", eval, eval)
		calcValues = append(calcValues, eval)
	}
	row.Cols = append(row.Cols, calcValues...)
	return row

	//indices := calcDef.Indices
	//var nCols []interface{}
	//added := false
	//for i, col := range row.Cols {
	//	if !indices.Contains(i) {
	//		nCols = append(nCols, col)
	//	} else {
	//		if !added {
	//			nCols = append(nCols, eval)
	//			added = true
	//		}
	//	}
	//}
	//return &DataRow{
	//	Cols:  nCols,
	//	Count: row.Count,
	//}
}

func hasFloatArgs(params govaluate.MapParameters) bool {
	for _, value := range params {
		switch value.(type) {
		case float64:
			return true
		default:
			return false
		}
	}
	return false
}

//func handleJsonOutput(csvFmt *CsvFormat, processor *LineProcessor) {
//	rows := make([]map[string]string, 0)
//	for _, line := range processor.Lines {
//		row := make(map[string]string)
//		for i, header := range processor.Header {
//			row[header] = line[i]
//		}
//		rows = append(rows, row)
//	}
//	buf, err := json.Marshal(rows)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Printf("%s\n", buf)
//}

func processJsonOutput(rows []DataRow, csvFmt *CsvFormat, headers []string, groupByCount int) {
	output := csvFmt.OutputDef
	fields := output.Fields
	levels := output.Levels
	if len(fields) == 0 {
		fields = calculateOutputFields(headers, levels, groupByCount)
	} else {
		levels = len(fields) - len(rows[0].Cols) //recalculate the levels
		if levels < 0 {
			fmt.Printf("Invalid JSON Fields, expected atleast %v\n", len(rows[0].Cols))
			return
		}
	}
	if levels == 0 {
		array := make([]map[string]interface{}, 0)
		for _, row := range rows {
			colMap := make(map[string]interface{})
			for i, col := range row.Cols {
				colMap[fields[i]] = col
			}
			array = append(array, colMap)
		}
		printJson(array)
	} else {
		outMap := make(map[string]map[string]interface{})
		for _, row := range rows {
			processJsonLevel(&row, 0, fields, outMap)
		}
		//unwrap Json
		printJson(unwrapJsonMap(0, levels, outMap))
	}
}

func calculateOutputFields(headers []string, levels int, keyCount int) []string {
	if levels == 0 {
		return headers
	}
	if levels > keyCount {
		fmt.Printf("ERROR: The level value of %d is invalid. Max allowed is %d\n", levels, keyCount)
		levels = keyCount
	}
	var nHeaders []string
	levelIdx := levels
	for i, header := range headers {
		nHeaders = append(nHeaders, header)
		if levelIdx > 0 {
			if levelIdx == 1 {
				nHeaders = append(nHeaders, header+"-group")
			} else {
				nHeaders = append(nHeaders, headers[i+1]+"s")
			}
		}
		levelIdx--
	}
	return nHeaders
}

func printJson(array []map[string]interface{}) {
	buf, err := json.Marshal(array)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", buf)
}

func unwrapJsonMap(level int, levels int, dataMap map[string]map[string]interface{}) []map[string]interface{} {
	var array []map[string]interface{}
	for _, valueMap := range dataMap {
		for key, value := range valueMap {
			switch value.(type) {
			case map[string]map[string]interface{}:
				valueMap[key] = unwrapJsonMap(level, levels, value.(map[string]map[string]interface{}))
			}
		}
		array = append(array, valueMap)
	}
	return array
}

func processJsonLevel(row *DataRow, level int, fields []string, dataMap map[string]map[string]interface{}) {
	levels := len(fields) - len(row.Cols)
	fieldIdx := level * 2
	fieldName := fields[fieldIdx]
	colName := row.Cols[level].(string)
	colMap, exists := dataMap[colName]
	if !exists {
		colMap = make(map[string]interface{})
		dataMap[colName] = colMap
	}
	colMap[fieldName] = colName
	fieldIdx++
	nFieldName := fields[fieldIdx]
	//fmt.Println("level=", level, "levels=", levels, "fieldName=", fieldName, "colName=", colName, "nextField=", nFieldName)
	if level == levels-1 {
		_, exists := colMap[nFieldName]
		if !exists {
			colMap[nFieldName] = make([]map[string]interface{}, 0)
		}
		fieldIdx++
		subArray := colMap[nFieldName].([]map[string]interface{})
		subMap := make(map[string]interface{})
		colIdx := level + 1

		for i := fieldIdx; i < len(fields); i++ {
			//fmt.Println("fieldIndex=", i, len(fields))
			//fmt.Println("colIndex=", colIdx, len(row.Cols))
			subMap[fields[i]] = row.Cols[colIdx]
			colIdx++
		}
		subArray = append(subArray, subMap)
		colMap[nFieldName] = subArray // i dont like this, need to use an object
	} else {
		nextColMap, exists := colMap[nFieldName]
		if !exists {
			nextColMap = make(map[string]map[string]interface{})
			colMap[nFieldName] = nextColMap
		}
		//fmt.Println("--------- NEXT ------------")
		processJsonLevel(row, level+1, fields, nextColMap.(map[string]map[string]interface{}))
	}
}

func toDataRows(processor *LineProcessor) []DataRow {
	var dataRows []DataRow
	for _, line := range processor.Lines {
		var nWords []interface{}
		for _, word := range line {
			nWords = append(nWords, word)
		}
		dataRows = append(dataRows, DataRow{Cols: nWords})
	}
	return dataRows
}

func convertAndApplySort(csvFmt *CsvFormat, rows []DataRow) []DataRow {
	firstRow := rows[0]
	sortDef := csvFmt.SortDef
	sortCols := sortDef.SortCols
	sortIndices := common.GetFilterItemIndices(common.IApplyRange(firstRow.Cols, sortCols))
	var dataRows []DataRow
	indexMap := make(map[int]bool)
	for _, index := range sortIndices {
		indexMap[index] = true
	}
	for _, row := range rows {
		var nWords []interface{}
		for i, word := range row.Cols {
			if _, exists := indexMap[i]; exists {
				nWords = append(nWords, ConvertIfNeeded(word))
			} else {
				nWords = append(nWords, word)
			}
		}
		dataRows = append(dataRows, DataRow{Cols: nWords})
	}
	rowSort := &DataRowSort{
		Rows:    dataRows,
		Indices: sortIndices,
		Desc:    sortDef.Desc,
	}
	sort.Sort(rowSort)
	return dataRows
}

func applySort(rows []DataRow, csvFmt *CsvFormat) {
	if len(rows) == 0 {
		return
	}
	sortDef := csvFmt.SortDef
	sortCols := sortDef.SortCols
	sortIndices := common.GetFilterItemIndices(common.IApplyRange(rows[0].Cols, sortCols))
	rowSort := &DataRowSort{
		Rows:    rows,
		Indices: sortIndices,
		Desc:    sortDef.Desc,
	}
	sort.Sort(rowSort)
}

func hasWholeOpr(csvFmt *CsvFormat) bool {
	if csvFmt.IsLMerge || csvFmt.MapRed != nil || HasNonCsvOutputFmt(csvFmt) || csvFmt.SortDef != nil {
		return true
	}
	return false
}

func HasNonCsvOutputFmt(csvFmt *CsvFormat) bool {
	return csvFmt.OutputDef != nil && csvFmt.OutputDef.Type != "csv"
}

func hasMrReducer(mapRed *GroupByDef) bool {
	return mapRed != nil
}

func parseInlineCommand(cmd string, cmdline string) []string {
	chars := []rune(cmdline)
	if len(cmd) == len(cmdline) {
		return []string{cmdline}
	}
	r := chars[len(cmd)]
	var sep []rune
	sep = append(sep, r)
	for i := len(cmd) + 1; i < len(chars); i++ {
		if chars[i] != r {
			break
		}
		sep = append(sep, chars[i])
	}
	return strings.Split(cmdline, string(sep))
}

func extractHeaderDef(arg string) *HeaderDef {
	parts := common.ParseSubCommandArg(arg)
	def := &HeaderDef{}
	def.Fields = common.ParseIndexStr(parts[0])
	return def
}

func extractCalcDef(arg string, csvFmt *CsvFormat) {
	def := CalcDef{}
	rawExpr := common.ParseExprStr(arg)
	modifiedExpr := rawExpr
	var indices common.IntSet
	for i := 0; i < 100; i++ {
		idxVar := fmt.Sprintf("[%d]", i)
		if strings.Index(rawExpr, idxVar) != -1 {
			indices.Add(i)
			modifiedExpr = strings.Replace(modifiedExpr, idxVar,
				fmt.Sprintf("col%d", i), 1)
		}
	}
	def.RawExpr = rawExpr
	def.ParsedExpr = modifiedExpr
	def.Indices = &indices
	expr, err := govaluate.NewEvaluableExpression(def.ParsedExpr)
	if err != nil {
		fmt.Printf("ERR: Expr eval failed %v", err)
		return
	}
	def.EvalExpr = expr
	csvFmt.CalcDefs = append(csvFmt.CalcDefs, def)
}

func processOutputArgs(command string, c *CsvFormat) {
	parts := parseInlineCommand("out", command)
	def := OutputDef{}
	def.Type = parts[1]
	for _, arg := range parts[2:] {
		if strings.Index(arg, "fields[") == 0 {
			fields := common.ParseSubCommandArg(arg)
			def.Fields = common.ParseIndexStr(fields[0])
		} else if strings.Index(arg, "levels:") != -1 {
			levels, err := strconv.Atoi(extractArg(arg, "levels:"))
			if err == nil {
				def.Levels = levels
			}
		}
	}
	c.OutputDef = &def
}

func processGroupArgs(command string, csvFmt *CsvFormat) {
	args := common.ParseSubCommandArg(command)
	mapRed := GroupByDef{}
	for _, part := range args {
		if strings.Index(part, "group[") != -1 {
			mapRed.ColIndices = extractCsvIndexArg(part)
		} else if strings.Index(part, "count") != -1 {
			mapRed.ShowCount = true
		}
	}
	csvFmt.MapRed = &mapRed
}

//func processMrArguments(command string, csvFmt *CsvFormat) {
//	mapRed := GroupByDef{}
//	parts := parseInlineCommand("mr", command)[1:]
//	for _, part := range parts {
//		if strings.Index(part, "sum") != -1 {
//			split := strings.Split(part, ":")
//			if len(split) == 1 {
//				mapRed.Sum = "col"
//			} else {
//				mapRed.Sum = split[1]
//			}
//		} else if strings.Index(part, "group[") != -1 {
//			split := common.ParseSubCommandArg(part)
//			mapRed.ColIndices = extractCsvIndexArg(split[0])
//		} else if strings.Index(part, "sort[") != -1 {
//			sortDef := extractSort(part)
//			mapRed.SortDef = sortDef
//		}
//	}
//	csvFmt.GroupByDef = &mapRed
//}

func extractSort(part string) *SortDef {
	split := common.ParseSubCommandArg(part)
	sortDef := &SortDef{}
	sortDef.SortCols = extractCsvIndexArg(split[0])
	if len(split) > 1 {
		if split[1] == "desc" {
			sortDef.Desc = true
		} else {
			sortDef.Desc = false
		}
	}
	return sortDef
}

func processFmtArguments(command string, csvFmt *CsvFormat) {
	parts := parseInlineCommand("fmt", command)
	colIndex, err := strconv.Atoi(strings.Replace(parts[1], "c", "", 1))
	if err != nil {
		panic("Invalid col index for formatting" + parts[1])
	}
	format := ColumnFormat{}
	for _, part := range parts[2:] {
		if strings.Index(part, "split:") == 0 {
			format.Split = extractDelim(part, "split:")
		} else if strings.Index(part, "merge:") == 0 {
			format.Merge = extractDelim(part, "merge:")
		} else if strings.Index(part, "wrap:") == 0 {
			format.Wrap = extractDelim(part, "wrap:")
		} else if strings.Index(part, "sfx:") == 0 {
			format.Suffix = extractArg(part, "sfx:")
		} else if strings.Index(part, "pfx:") == 0 {
			format.Prefix = extractArg(part, "pfx:")
		} else if strings.Index(part, "col[") == 0 {
			format.ColExt = extractCsvIndexArg(part)
		} else if strings.Index(part, "ncol[") == 0 {
			format.ColExt = extractCsvIndexArg(part)
			format.ColExt.Exclude = true
		}
	}
	fmtMap := csvFmt.ColFmtMap
	if fmtMap == nil {
		fmtMap = make(map[int][]ColumnFormat)
		csvFmt.ColFmtMap = fmtMap
	}
	existing, exists := fmtMap[colIndex]
	if !exists {
		fmtMap[colIndex] = []ColumnFormat{format}
	} else {
		fmtMap[colIndex] = append(existing, format)
	}
}

func extractArg(arg string, prefix string) string {
	return strings.Replace(arg, prefix, "", 1)
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

func extractCsv(words []string, ext *common.IntRange, fmtMap map[int][]ColumnFormat) []string {
	ftrWords := common.ApplyRange(words, ext)
	var vals []string
	for _, ftrWord := range *ftrWords {
		word := processWord(ftrWord, fmtMap)
		vals = append(vals, word)
	}
	return vals
}

func processWord(ftrWord common.FilterStr, fmtMap map[int][]ColumnFormat) string {
	if fmtMap == nil {
		return ftrWord.Str
	}
	colFormats, exists := fmtMap[ftrWord.Index]
	if !exists || len(colFormats) <= 0 {
		return ftrWord.Str
	}
	word := ftrWord.Str
	for _, colFormat := range colFormats {
		if colFormat.Split != "" {
			parts := strings.Split(word, colFormat.Split)
			if colFormat.ColExt != nil {
				parts = extractCsv(parts, colFormat.ColExt, nil)
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
