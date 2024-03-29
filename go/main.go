package main

import (
	"fmt"
	"github.com/abeytom/utilbox/gcloud"
	"github.com/abeytom/utilbox/k8"
	"github.com/abeytom/utilbox/utils"
	"os"
)

func main() {
	args := os.Args
	if args[1] == "utils" {
		utils.Execute(args[1:])
	} else if args[1] == "run" {
		utils.ExecuteCommand(args[1:])
	} else if args[1] == "k8" {
		k8.Execute(args)
	} else if args[1] == "fmt" {
		utils.Format(args[1:])
	} else if args[1] == "gcloud_art" {
		gcloud.Execute(args[2:])
	} else if args[1] == "jsonLog2Txt" {
		utils.JsonLog2Txt()
	} else if args[1] == "tok" {
		utils.BearerToken(args[1:])
	} else if args[1] == "curl" {
		utils.Curl(args[2:])
	} else if args[1] == "json_parse" {
		utils.JsonParse(args[2:])
	} else if args[1] == "json_parse_line" {
		utils.JsonParseLine(args[2:])
	} else if args[1] == "csv_parse" {
		utils.CsvParse(args[2:])
	} else if args[1] == "yaml_parse" {
		utils.YamlParse(args[2:])
	} else if args[1] == "vbox" {
		utils.VbExec(args[2:])
	} else if args[1] == "awx" {
		utils.AwsExec(args[2:])
	} else if args[1] == "gx" {
		utils.GxExec(args[2:])
	} else if args[1] == "regex" {
		utils.RegexExtract(args[2:])
	} else if args[1] == "hist" {
		utils.ListHistory(args[2:])
	} else {
		fmt.Printf("Unknown command %s\n", args)
	}
}
