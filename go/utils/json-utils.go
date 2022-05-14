package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Knetic/govaluate"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
)

/*
cat /Users/atom/tmp/pods.json | jp keys

cat /Users/atom/github/abeytom/utilbox/go/resources/pods.json | jp keys[items.metadata.name,items.spec.containers.args]  out..table

cat /Users/atom/github/abeytom/utilbox/go/resources/pods.json | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..table
cat /Users/atom/github/abeytom/utilbox/go/resources/pods.json | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..table head[name,ns,hostIp,podIp]


cat /Users/atom/github/abeytom/utilbox/go/resources/pods.json | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..table head[name,ns,hostIp,podIp] | csv row[1:] group[0]:count out..table tr..c0..split:-..merge:-..ncol[-1] sort[4]:desc

*/
import (
	"fmt"
	"github.com/abeytom/utilbox/common"
	"strings"
)

type TreeNode struct {
	Map    map[string]*TreeNode
	Parent *TreeNode
	Key    string
	Leaf   bool
}

func NewTreeNode() *TreeNode {
	return &TreeNode{Map: make(map[string]*TreeNode)}
}

func (t *TreeNode) Add(segments []string) {
	key := segments[0]
	node, exists := t.Map[key]
	if !exists {
		node = NewTreeNode()
		node.Parent = t
		node.Key = key
		t.Map[key] = node
	}
	if len(segments) > 1 {
		node.Add(segments[1:])
	} else {
		node.Leaf = true
	}
}

func (t *TreeNode) FullKey() string {
	if t.Parent == nil {
		return t.Key
	}
	return appendKey(t.Parent.FullKey(), t.Key)
}

func readStdIn() []byte {
	stat, _ := os.Stdin.Stat()
	newLineByte := []byte("\n")[0]
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
			stdin = append(stdin, newLineByte)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		return stdin
	} else {
		return nil
	}
}

func readStdIn2(lineCb func(line []byte)) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lineCb(scanner.Bytes())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
}

func JsonParseLine(args []string) {
	var wExpr *ExprWrap
	csvFmt := &CsvFormat{
		ColExt:      &common.IntRange{},
		RowExt:      &common.IntRange{},
		Split:       "space+",
		Merge:       " ",
		LMerge:      ",",
		Wrap:        "",
		IsLMerge:    false,
		NoHeaderOut: true,
	}
	doParseCsvArgs(args, csvFmt)
	hasFilter := csvFmt.Filter != nil && csvFmt.Filter.Expr != nil
	if hasFilter {
		wExpr = NewExprWrap(csvFmt.Filter.Expr)
	}
	if csvFmt.KeyDef == nil {
		if !hasFilter {
			log.Fatal(errors.New("No keys"))
		}
		applyFilter(csvFmt)
		return
	}
	printKeys := len(csvFmt.KeyDef.Fields) == 0
	keyMap := make(map[string]bool)

	var cb = func(line []byte) {
		array := parseJsonBytes(line)
		//print the line if it is not a json
		if array == nil {
			if !printKeys {
				fmt.Println(string(line))
			}
			return
		}
		if printKeys {
			keys := JsonKeys(array)
			for _, key := range keys {
				keyMap[key.Key] = true
			}
		} else {
			if !applyFilter2(wExpr, array) {
				return
			}
			keys := csvFmt.KeyDef.Fields
			rows := Flatten(array, keys)
			processOutput(csvFmt, &DataRows{
				DataRows:     rows,
				Headers:      keys,
				GroupByCount: 0,
				Converted:    false,
			})
		}
	}
	readStdIn2(cb)
	if printKeys {
		keys := make([]string, len(keyMap))
		i := 0
		for k := range keyMap {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("%v\n", key)
		}
	}
}

func applyFilter2(wExpr *ExprWrap, array []map[string]interface{}) bool {
	if wExpr == nil {
		return true
	}
	params := make(map[string]interface{})
	for _, key := range wExpr.keys {
		value := getValueForKeyFromArray(array, key)
		params[key] = wExpr.convertValue(key, value)
	}
	evaluate, err := wExpr.expr.Evaluate(params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while eval '%v' with params '%+v'. The error is '%v'",
			wExpr.expr, params, err)
		return false
	}
	switch evaluate.(type) {
	case bool:
		if evaluate.(bool) {
			return true
		}
	}
	return false
}

func getValueForKeyFromArray(array []map[string]interface{}, key string) interface{} {
	segments := splitKey(key)
	if len(array) == 0 {
		return ""
	}
	if len(array) == 1 {
		return getValueForKeyFromMap(array[0], segments)
	}
	values := make([]interface{}, len(array))
	for i, json := range array {
		values[i] = getValueForKeyFromMap(json, segments)
	}
	return values
}

func getValueForKeyFromMap(json map[string]interface{}, segments []string) interface{} {
	value, exists := json[segments[0]]
	if !exists {
		return ""
	}
	if len(segments) == 1 {
		return value
	}
	switch value.(type) {
	case []interface{}:
		var subVals []interface{}
		for _, e := range value.([]interface{}) {
			switch e.(type) {
			case map[string]interface{}:
				subVal := getValueForKeyFromMap(e.(map[string]interface{}), segments[1:])
				subVals = append(subVals, subVal)
			}
		}
		return subVals
	case map[string]interface{}:
		return getValueForKeyFromMap(value.(map[string]interface{}), segments[1:])
	}
	return ""
}

func applyFilter(csvFmt *CsvFormat) {
	expr := csvFmt.Filter.Expr
	wExpr := NewExprWrap(expr)
	if len(wExpr.keys) == 0 {
		log.Fatalf("Invalid Expr '%v'. Atleast one variable is expected", csvFmt.Filter.ExprStr)
	}
	var cb = func(line []byte) {
		array := parseJsonBytes(line)
		if applyFilter2(wExpr, array) {
			fmt.Println(string(line))
		}
	}
	readStdIn2(cb)
}

func JsonParse(args []string) {
	jsonBytes := readStdIn()
	if jsonBytes == nil {
		log.Fatal(errors.New("there is no data to read from STDIN"))
	}
	array := parseJsonBytes(jsonBytes)
	if array == nil {
		log.Fatalf("Unsupported JSON %s", string(jsonBytes))
	}
	csvFmt := &CsvFormat{
		ColExt:    &common.IntRange{},
		RowExt:    &common.IntRange{},
		Split:     "space+",
		Merge:     "csv",
		LMerge:    ",",
		Wrap:      "",
		OutputDef: &OutputDef{Type: "table"},
		IsLMerge:  false,
	}
	doParseCsvArgs(args, csvFmt)
	if csvFmt.KeyDef != nil {
		if len(csvFmt.KeyDef.Fields) == 0 {
			keys := JsonKeys(array)
			for _, key := range keys {
				if strings.Index(key.Key, "\\.") != -1 {
					fmt.Printf("'%v'\n", key.Key)
				} else {
					fmt.Printf("%v\n", key.Key)
				}
			}
		} else {
			keys := csvFmt.KeyDef.Fields
			rows := Flatten(array, keys)
			processOutput(csvFmt, &DataRows{
				DataRows:     rows,
				Headers:      keys,
				GroupByCount: 0,
				Converted:    false,
			})
		}
	}
}

func parseJsonBytes(jsonBytes []byte) []map[string]interface{} {
	x := bytes.TrimLeft(jsonBytes, " \t\r\n")
	isArray := len(x) > 0 && x[0] == '['
	isObject := len(x) > 0 && x[0] == '{'

	var array []map[string]interface{}
	if isObject {
		var jsonMap map[string]interface{}
		err := json.Unmarshal(jsonBytes, &jsonMap)
		if err != nil {
			log.Printf("Error while marshalling json into map. %v\n", err)
			return nil
		}
		array = append(array, jsonMap)
	} else if isArray {
		err := json.Unmarshal(jsonBytes, &array)
		if err != nil {
			log.Printf("Error while marshalling json into array. %v\n", err)
			return nil
		}
	} else {
		return nil
	}
	return array
}

func Flatten(array []map[string]interface{}, keys []string) []DataRow {
	root := NewTreeNode()
	for _, key := range keys {
		segments := splitKey(key)
		root.Add(segments)
	}
	var rows []DataRow
	for _, json := range array {
		result := make(map[string][]interface{})
		flatten(json, root, 0, result)
		rows = processFlattenedResults(result, keys, rows)
	}
	return rows
}

func convertValuesToStringSet(values []interface{}) *common.StringList {
	comparable := true
	set := &common.StringList{}
L:
	for _, value := range values {
		vref := reflect.ValueOf(value)
		switch vref.Kind() {
		case reflect.Slice:
			comparable = false
			break L
		case reflect.Map:
			comparable = false
			break L
		}
		set.Add(fmt.Sprintf("%v", value))
	}
	if comparable {
		return set
	}
	return nil
}

func processFlattenedResults(result map[string][]interface{}, keys []string, rows []DataRow) []DataRow {
	if len(result) == 0 {
		return rows
	}
	//todo fixme this needs to be re-written
	//buggy + buggy + buggy
	resultRows, exists := result["_$result$_"]
	if !exists {
		for _, key := range keys {
			values, valExists := result[key]
			if !valExists {
				return rows
			}
			size := len(values)
			if size == 0 {
				return rows
			} else if size == 1 {
				return append(rows, DataRow{Cols: []interface{}{values[0]}})
			}
			set := convertValuesToStringSet(values)
			if set == nil {
				for _, value := range values {
					rows = append(rows, DataRow{Cols: []interface{}{value}})
				}
			} else {
				for _, value := range set.Values() {
					rows = append(rows, DataRow{Cols: []interface{}{value}})
				}
			}
		}

		return rows
	}

	if exists {
		for _, resultsVal := range resultRows {
			resultMap := resultsVal.(map[string][]interface{})
			rowCols := make([]interface{}, len(keys))
			for i, key := range keys {
				//fmt.Println(key)
				values, exists2 := resultMap[key]
				if !exists2 || len(values) == 0 {
					rowCols[i] = ""
				} else if len(values) == 1 {
					rowCols[i] = values[0]
				} else if len(values) > 1 {
					set := convertValuesToStringSet(values)
					if set != nil {
						rowCols[i] = set
					} else {
						rowCols[i] = values
					}
				} else {
					rowCols[i] = ""
				}
			}
			rows = append(rows, DataRow{Cols: rowCols})
		}
		return rows
	}
	return rows
}

func flatten(json map[string]interface{}, root *TreeNode, depth int, inResult map[string][]interface{}) {
	if depth == 0 {
		if len(root.Map) > 1 {
			depth = 1
		}
	} else {
		depth++
	}
	result := inResult
	if depth > 0 {
		result = make(map[string][]interface{})
	}
	for key, value := range root.Map {
		v, exists := json[key]
		if !exists {
			continue
		}
		switch v.(type) {
		case []interface{}:
			for _, e := range v.([]interface{}) {
				switch e.(type) {
				case map[string]interface{}:
					if len(value.Map) > 0 {
						flatten(e.(map[string]interface{}), value, depth, result)
						if value.Leaf {
							appendResult(value.FullKey(), e, result)
						}
					} else {
						appendResult(value.FullKey(), e, result)
					}
				default:
					appendResult(value.FullKey(), e, result)
				}
			}
		case map[string]interface{}:
			if len(value.Map) > 0 {
				flatten(v.(map[string]interface{}), value, depth, result)
				if value.Leaf {
					appendResult(value.FullKey(), v, result)
				}
			} else {
				appendResult(value.FullKey(), v, result)
			}
		default:
			appendResult(value.FullKey(), v, result)
		}
	}
	if depth > 0 {
		if depth > 1 {
			for k, v := range result {
				for _, val := range v {
					appendResult(k, val, inResult)
				}
			}
		}
		if depth == 1 {
			appendResult("_$result$_", result, inResult)
		}
	}
}

func appendResult(key string, value interface{}, result map[string][]interface{}) {
	values, exists := result[key]
	if exists {
		result[key] = append(values, value)
	} else {
		result[key] = []interface{}{value}
	}
}

func JsonKeys(jsonArray []map[string]interface{}) []common.CountMapEntry {
	countMap := common.NewCountMap()
	for _, jsonMap := range jsonArray {
		jsonKeys("", jsonMap, countMap)
	}
	return countMap.Entries()
}

func jsonKeys(prefix string, json map[string]interface{}, countMap *common.CountMap) {
	for k, v := range json {
		full := appendKey(prefix, k)
		countMap.Add(full)
		//fmt.Println("Type", full, reflect.TypeOf(v).String())
		switch v.(type) {
		case []interface{}:
			for _, e := range v.([]interface{}) {
				itemPrefix := full
				countMap.Add(itemPrefix)
				switch e.(type) {
				case map[string]interface{}:
					jsonKeys(itemPrefix, e.(map[string]interface{}), countMap)
				}
			}
		case map[string]interface{}:
			jsonKeys(full, v.(map[string]interface{}), countMap)
		}
	}
}

func appendKey(prefix string, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + strings.ReplaceAll(key, ".", "\\.")
}

func splitKey(key string) []string {
	chars := []rune(key)
	prevIdx := 0
	var segments []string
	for i := 0; i < len(chars); i++ {
		if i > 0 && chars[i] == '.' && chars[i-1] != '\\' {
			segment := string(chars[prevIdx:i])
			segments = append(segments, strings.ReplaceAll(segment, "\\.", "."))
			prevIdx = i + 1
		}
	}
	segment := string(chars[prevIdx:])
	segments = append(segments, strings.ReplaceAll(segment, "\\.", "."))
	return segments
}

type ExprWrap struct {
	expr     *govaluate.EvaluableExpression
	keys     []string
	valueMap map[string]govaluate.ExpressionToken
}

func (e *ExprWrap) convertValue(key string, value interface{}) interface{} {
	token, exists := e.valueMap[key]
	if !exists {
		return value
	}
	switch token.Kind {
	case govaluate.NUMERIC:
		switch value.(type) {
		case string:
			int64Val, err := strconv.ParseInt(value.(string), 10, 64)
			if err == nil {
				return int64Val
			}
			float64Val, err := strconv.ParseFloat(value.(string), 64)
			if err == nil {
				return float64Val
			}
		}
	case govaluate.STRING:
		switch value.(type) {
		case string:
			return value
		default:
			return fmt.Sprintf("%v", value)
		}
	}
	return value
}

func NewExprWrap(expr *govaluate.EvaluableExpression) *ExprWrap {
	tokens := expr.Tokens()
	valueMap := make(map[string]govaluate.ExpressionToken)
	var keys []string
	for i := 0; i < len(tokens)-2; i++ {
		token := tokens[i]
		if token.Kind == govaluate.VARIABLE {
			if tokens[i+1].Kind == govaluate.COMPARATOR {
				valueMap[token.Value.(string)] = tokens[i+2]
				i += 2
			}
			keys = append(keys, token.Value.(string))
		}

	}
	return &ExprWrap{expr: expr, keys: keys, valueMap: valueMap}
}
