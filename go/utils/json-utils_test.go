package utils

import (
	"testing"
)

func TestReadJson(t *testing.T) {
	//jsonBytes, _ := ioutil.ReadFile("/Users/atom/tmp/pods.json")
	//var jsonMap map[string]interface{}
	//json.Unmarshal(jsonBytes, &jsonMap)
	//entries := JsonKeys(jsonMap)
	//for _, key := range entries {
	//	fmt.Println(key.Key + ": " + strconv.Itoa(key.Count))
	//}
	//keys := []string{"items.metadata.name", "items.metadata.namespace", "items.metadata.labels.app", "items.spec.containers.image","items.status.hostIP","items.status.podIP"}
	//Flatten(jsonMap, keys)

	//keys2 := []string{"items.metadata.name", "items.metadata.namespace"}
	//Flatten(jsonMap, keys2)

	//keys := []string{"items.metadata.name"}
	//keys := []string{"items.metadata.name","items.spec.initContainers.args"}
	//Flatten(jsonMap, keys3)

	//rows := Flatten(jsonMap, keys)
	//csvFmt := &CsvFormat{OutputDef: &OutputDef{Type: "table"}}
	//processTableOutput(rows, csvFmt, keys)

	//fmt.Println(jsonMap)
	//for k, v := range jsonMap {
	//	fmt.Printf("k=%v, value = %+v\n", k, v)
	//}
}
