package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/abeytom/utilbox/common"
	"gopkg.in/yaml.v2"
	"log"
	"strings"
)

func YamlParse(args []string) {
	yamlBytes := readStdIn()
	if yamlBytes == nil {
		log.Fatal(errors.New("there is no data to read from STDIN"))
	}
	x := bytes.TrimLeft(yamlBytes, " \t\r\n")
	isArray := len(x) > 0 && x[0] == '-'
	var array []map[interface{}]interface{}
	if isArray {
		err := yaml.Unmarshal(yamlBytes, &array)
		if err != nil {
			log.Printf("Error while marshalling YAML into array. The error is [%v]\n", err)
		}
	} else {
		var jsonMap map[interface{}]interface{}
		err := yaml.Unmarshal(yamlBytes, &jsonMap)
		if err != nil {
			log.Printf("Error while marshalling YAML into map. The error is [%v]\n", err)
		}
		array = append(array, jsonMap)
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
			keys := getYamlKeys(array)
			for _, key := range keys {
				fmt.Printf("%v\n", key.Key)
			}
		} else {
			keys := csvFmt.KeyDef.Fields
			rows := flattenYaml(array, keys)
			processOutput(csvFmt, &DataRows{
				DataRows:     rows,
				Headers:      keys,
				GroupByCount: 0,
				Converted:    false,
			})
		}
	}
}

func getYamlKeys(jsonArray []map[interface{}]interface{}) []common.CountMapEntry {
	countMap := common.NewCountMap()
	for _, jsonMap := range jsonArray {
		doGetYamlKeys("", jsonMap, countMap)
	}
	return countMap.Entries()
}

func doGetYamlKeys(prefix string, json map[interface{}]interface{}, countMap *common.CountMap) {
	for k, v := range json {
		full := appendYamlKey(prefix, k)
		countMap.Add(full)
		//fmt.Println("Type", full, reflect.TypeOf(v).String())
		switch v.(type) {
		case []interface{}:
			for _, e := range v.([]interface{}) {
				itemPrefix := full
				countMap.Add(itemPrefix)
				switch e.(type) {
				case map[interface{}]interface{}:
					doGetYamlKeys(itemPrefix, e.(map[interface{}]interface{}), countMap)
				}
			}
		case map[interface{}]interface{}:
			doGetYamlKeys(full, v.(map[interface{}]interface{}), countMap)
		}
	}
}

func appendYamlKey(prefix string, key0 interface{}) string {
	key := fmt.Sprintf("%s", key0)
	if prefix == "" {
		return key
	}
	return prefix + "." + strings.ReplaceAll(key, ".", "\\.")
}

func flattenYaml(yamlArray []map[interface{}]interface{}, keys []string) []DataRow {
	root := NewTreeNode()
	for _, key := range keys {
		segments := splitKey(key)
		root.Add(segments)
	}
	var rows []DataRow
	for _, yamlMap := range yamlArray {
		result := make(map[interface{}][]interface{})
		flattenYaml2(yamlMap, root, 0, result)
		rows = processFlattenedYamlResults(result, keys, rows)
	}
	return rows
}

func processFlattenedYamlResults(result map[interface{}][]interface{}, keys []string, rows []DataRow) []DataRow {
	if len(keys) == 1 {
		if len(result) == 0 {
			return rows
		}
		values, exists := result[keys[0]]
		if !exists {
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
		return rows
	}

	resultRows, exists := result["result"]
	if exists {
		for _, resultsVal := range resultRows {
			resultMap := resultsVal.(map[interface{}][]interface{})
			rowCols := make([]interface{}, len(keys))
			for i, key := range keys {
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

func flattenYaml2(json map[interface{}]interface{}, root *TreeNode, depth int, inResult map[interface{}][]interface{}) {
	if depth == 0 {
		if len(root.Map) > 1 {
			depth = 1
		}
	} else {
		depth++
	}
	result := inResult
	if depth > 0 {
		result = make(map[interface{}][]interface{})
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
				case map[interface{}]interface{}:
					if len(value.Map) > 0 {
						flattenYaml2(e.(map[interface{}]interface{}), value, depth, result)
					} else {
						appendYamlResult(value.FullKey(), e, result)
					}
				default:
					appendYamlResult(value.FullKey(), e, result)
				}
			}
		case map[interface{}]interface{}:
			if len(value.Map) > 0 {
				flattenYaml2(v.(map[interface{}]interface{}), value, depth, result)
			} else {
				appendYamlResult(value.FullKey(), v, result)
			}
		default:
			appendYamlResult(value.FullKey(), v, result)
		}
	}
	if depth > 0 {
		if depth > 1 {
			for k, v := range result {
				for _, val := range v {
					appendYamlResult(fmt.Sprintf("%v", k), val, inResult)
				}
			}
		}
		if depth == 1 {
			appendYamlResult("result", result, inResult)
		}
	}
}

func appendYamlResult(key string, value interface{}, result map[interface{}][]interface{}) {
	values, exists := result[key]
	if exists {
		result[key] = append(values, value)
	} else {
		result[key] = []interface{}{value}
	}
}
