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
	//rows := utils.Flatten(array, []string{"pods","pods.name","pods.name2"})
	//rows := utils.Flatten(array, []string{"pods.containers","pods.containers.name","pods.containers.name2"})
	rows := utils.Flatten(array, []string{"pods","pods.containers","pods.containers.name","pods.containers.name2"})
	//rows := utils.Flatten(array, []string{"pods.fname","pods.name"})
	fmt.Println("ROWS",len(rows))
	for _, row := range rows {
		fmt.Println(len(row.Cols))
	}
}
