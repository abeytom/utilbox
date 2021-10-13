package tests

import (
	"encoding/json"
	"fmt"
	"github.com/abeytom/utilbox/utils"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestJsonValueMerge(t *testing.T) {
	file, err := os.Open("./custom.json")
	if err != nil {
		log.Fatal(err)
	}
	bytes, _ := ioutil.ReadAll(file)
	var jsonMap map[string]interface{}
	json.Unmarshal(bytes, &jsonMap)
	array := make([]map[string]interface{}, 1)
	array[0] = jsonMap
	rows := utils.Flatten(array, []string{"pods.name", "pods.containers.name"})
	fmt.Println("ROWS")
	for _, row := range rows {
		fmt.Println(row)
	}
}
