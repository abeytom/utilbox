package main

import (
	"fmt"
	"github.com/abeytom/cmdline-utils/gcloud"
	"github.com/abeytom/cmdline-utils/k8"
	"github.com/abeytom/cmdline-utils/utils"
	"os"
)

func main() {
	args := os.Args
	if args[1] == "utils" {
		utils.Execute(args[1:])
	} else if args[1] == "k8" {
		k8.Execute(args[1:])
	} else if args[1] == "fmt" {
		utils.Format(args[1:])
	} else if args[1] == "gcloud_art" {
		gcloud.Execute(args[2:])
	} else if args[1] == "jsonLog2Txt" {
		utils.JsonLog2Txt()
	} else if args[1] == "tok" {
		utils.BearerToken(args[1:])
	} else {
		fmt.Printf("Unknown command %s\n", args)
	}
}
