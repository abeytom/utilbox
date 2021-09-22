package utils

import (
	"encoding/base64"
	"log"
	"os"
	"os/exec"
)

func Curl(args []string) {
	var nArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-json" {
			nArgs = append(nArgs, "-H", "Content-Type: application/json",
				"-H", "Accept: application/json")
		} else if arg == "-bearer" {
			i++
			nArgs = append(nArgs, appendBearer(args[i])...)
		} else if arg == "-basic" {
			i++
			nArgs = append(nArgs, appendBasic(args[i])...)
		} else {
			nArgs = append(nArgs, arg)
		}
	}
	//var out bytes.Buffer
	//command.Stdout = &out

	command := exec.Command("curl", nArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func appendBearer(key string) []string {
	value, err := GetKeyValue(key)
	var token string
	if err != nil {
		token = key
	} else {
		token = value
	}
	return []string{"-H", "Authorization: Bearer " + token}
}

func appendBasic(key string) []string {
	value, err := GetKeyValue(key)
	var userPwd string
	if err != nil {
		userPwd = key
	} else {
		userPwd = value
	}
	base64.StdEncoding.EncodeToString([]byte(userPwd))
	return []string{"-H", "Authorization: Basic " + userPwd}
}
