package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
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

func JsonParse(args []string) {
	jsonBytes := readStdIn()
	if jsonBytes == nil{
		log.Fatal(errors.New("there is no data to read from STDIN"))
	}
	x := bytes.TrimLeft(jsonBytes, " \t\r\n")
	isArray := len(x) > 0 && x[0] == '['
	isObject := len(x) > 0 && x[0] == '{'

	var array []map[string]interface{}
	if isObject {
		var jsonMap map[string]interface{}
		err := json.Unmarshal(jsonBytes, &jsonMap)
		if err != nil {
			log.Printf("Error while marshalling json into map. %v\n", err)
		}
		array = append(array, jsonMap)
	} else if isArray {
		err := json.Unmarshal(jsonBytes, &array)
		if err != nil {
			log.Printf("Error while marshalling json into array. %v\n", err)
		}
	} else {
		log.Fatalf("Unsupported JSON %s", string(jsonBytes))
	}

	csvFmt := parseCsvArgs(args)
	if csvFmt.KeyDef != nil {
		if len(csvFmt.KeyDef.Fields) == 0 {
			keys := JsonKeys(array)
			for _, key := range keys {
				fmt.Printf("%v\n", key.Key)
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

func processFlattenedResults(result map[string][]interface{}, keys []string, rows []DataRow) []DataRow {
	if len(keys) == 1 {
		if len(result) == 0 {
			return rows
		}
		values, exists := result[keys[0]]
		if !exists {
			return rows
		}
		set := &common.StringSet{}
		for _, value := range values {
			set.Add(fmt.Sprintf("%v", value))
		}
		setVals := set.Values()
		for _, value := range setVals {
			rows = append(rows, DataRow{Cols: []interface{}{value}})
		}
		return rows
	}

	resultRows, exists := result["result"]
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
					strVals := make([]string, len(values))
					for i2, value := range values {
						strVals[i2] = fmt.Sprintf("%v", value)
					}
					rowCols[i] = common.NewStringSet(strVals)
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
					flatten(e.(map[string]interface{}), value, depth, result)
				default:
					appendResult(value.FullKey(), e, result)
				}
			}
		case map[string]interface{}:
			flatten(v.(map[string]interface{}), value, depth, result)
		default:
			appendResult(value.FullKey(), v, result)
		}
	}
	if depth > 0 {
		if depth > 1 {
			for k, v := range result {
				inResult[k] = v
			}
		}
		if depth == 1 {
			appendResult("result", result, inResult)
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
