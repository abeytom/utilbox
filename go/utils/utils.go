package utils

import (
	"fmt"
	"github.com/abeytom/utilbox/common"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)
import "encoding/json"

type Conf struct {
	Paths   map[string]string `json:"paths"`
	Aliases map[string]string `json:"aliases"`
	Tokens  map[string]string `json:"tokens"`
}

/**
bk: add list get exe path pbc

bk add path alias /path/some/path
bk add cmd alias "ls -al"
bk get path alias
bk get cmd alias
bk get alias
bk list cmd
bk list path
bk exec cmdAlias
bk exec cmdAlias args1 arg2
bk exec cmdAlias `bk get path pathAlias` eg  [bk exec ll `bk get path go`]
. bk path pathAlias
open `bk get istore` [open in the pathAlias in finder]
bk pbc path alias | bk pbc cmd alias | bk pbc alias
*/
func Execute(args []string) {
	baseDir := os.Getenv("UTILBOX_PATH")
	if baseDir == "" {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		baseDir = filepath.Join(user.HomeDir, ".utilbox")
		//fmt.Printf("ERR:ENV_VAR_NOT_SET [CMDLINE_UTILS_PATH]")
		//os.Exit(1)
	}
	cmd := args[1]
	if cmd == "add" {
		subCmd := args[2]
		conf := getConf(baseDir)
		alias := args[3]
		if subCmd == "path" {
			location := args[4]
			paths := conf.Paths
			if path, ok := paths[alias]; !ok || (len(args) >= 6 && args[5] == "-f") {
				paths[alias] = location
				writeJson(conf, baseDir)
				fmt.Printf("INFO:PATH_ADDED [%s=%s]", alias, location)
			} else {
				fmt.Printf("ERR:PATH_ALIAS_EXISTS [%s=%s]", alias, path)
			}
		} else if subCmd == "cmd" {
			fullCmd := args[4]
			aliases := conf.Aliases
			if path, ok := aliases[alias]; !ok || (len(args) >= 6 && args[5] == "-f") {
				aliases[alias] = fullCmd
				writeJson(conf, baseDir)
				fmt.Printf("INFO:COMMAND_ADDED [%s=%s]", alias, fullCmd)
			} else {
				fmt.Printf("ERR:COMMAND_ALIAS_EXISTS [%s=%s]", alias, path)
			}
		}
	} else if cmd == "get" || cmd == "pbc" {
		conf := getConf(baseDir)
		subCmd := args[2]
		var mapVals map[string]string
		var alias string
		if subCmd == "cmd" {
			alias = args[3]
			mapVals = conf.Aliases
		} else if subCmd == "path" {
			alias = args[3]
			mapVals = conf.Paths
		} else {
			alias = args[2]
			mapVals = mergeMaps(conf.Aliases, conf.Paths)
		}
		if val, ok := mapVals[alias]; ok {
			fmt.Print(val)
		} else {
			fmt.Printf("ERR:INVALID_ALIAS [%s]", alias)
		}
	} else if cmd == "list" {
		subCmd := args[2]
		conf := getConf(baseDir)
		var mapVals map[string]string
		if subCmd == "cmd" || subCmd == "cmds" {
			mapVals = conf.Aliases
		} else if subCmd == "kv" {
			mapVals = conf.Tokens
		} else { //catch all
			mapVals = conf.Paths
		}
		keys := *getSortedMapKeys(&mapVals)
		fmt.Printf("INFO: LIST \n\n")
		for index, k := range keys {
			fmt.Printf("%d. %s=%s\n", index, k, mapVals[k])
		}
		fmt.Printf("\n")
	} else if cmd == "exec" {
		alias := args[2]
		cmdArgs := args[3:]
		conf := getConf(baseDir)
		aliases := conf.Aliases
		if path, ok := aliases[alias]; ok {
			allArgs := append([]string{path}, cmdArgs...)
			fmt.Printf("%s", strings.Join(allArgs, " "))
		} else {
			value, errStr := lookupValueByKeyIndex(alias, aliases, "COMMAND")
			if errStr != "" {
				fmt.Print(errStr)
			} else {
				allArgs := append([]string{value}, cmdArgs...)
				fmt.Printf("%s", strings.Join(allArgs, " "))
			}
		}
		//} else if cmd == "path" {
		//	// This and get works similar for most part.
		//	conf := getConf(baseDir)
		//	paths := conf.Paths
		//	pathAlias := args[2]
		//	if path, ok := paths[pathAlias]; ok {
		//		fmt.Printf("%s", getSubPaths(path, args[3:]))
		//	} else {
		//		value, errStr := lookupValueByKeyIndex(pathAlias, paths, "PATH")
		//		if errStr != "" {
		//			fmt.Print(errStr)
		//		} else {
		//			fmt.Printf("%s", getSubPaths(value, args[3:]))
		//		}
		//	}
	} else if cmd == "csv" {
		HandleCsv(args[2:])
	} else {
		fmt.Printf("ERR:UNKNOWN_COMMAND [%s]", cmd)
	}
}

func lookupValueByKeyIndex(indexStr string, kvMap map[string]string, typeStr string) (string, string) {
	cmdIndex := common.StrToInt(indexStr, -1)
	if cmdIndex >= 0 {
		keys := *getSortedMapKeys(&kvMap)
		if len(keys) > cmdIndex {
			return kvMap[keys[cmdIndex]], ""
		} else {
			return "", fmt.Sprintf("ERR:INVALID_%s_INDEX [%s]", typeStr, indexStr)
		}
	} else {
		return "", fmt.Sprintf("ERR:INVALID_%s_ALIAS [%s]", typeStr, indexStr)
	}
}

func getSortedMapKeys(mp *map[string]string) *[]string {
	m := *mp
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return &keys
}

func getSubPaths(base string, paths []string) string {
	if len(paths) > 0 {
		//files, err := ioutil.ReadDir(base)
		//if err != nil {
		//
		//}
	}
	return base
}

func writeJson(conf *Conf, baseDir string) {
	file, _ := json.MarshalIndent(conf, "", "  ")
	jsonFile := getJsonFilePath(baseDir)
	dir := filepath.Dir(jsonFile)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	ioutil.WriteFile(jsonFile, file, 0644)
}

func getConf(baseDir string) *Conf {
	jsonFile := getJsonFilePath(baseDir)
	_, statErr := os.Stat(jsonFile)
	if os.IsNotExist(statErr) {
		createConfJson(baseDir)
	}
	jsonBytes, readErr := ioutil.ReadFile(jsonFile)
	if readErr != nil {
		fmt.Printf("there was an error reading the file %s, %s", jsonFile, readErr)
		return nil
	}
	var conf Conf
	err := json.Unmarshal(jsonBytes, &conf)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		return nil
	}
	return &conf
}

func createConfJson(baseDir string) {
	conf := Conf{
		Paths:   map[string]string{},
		Aliases: map[string]string{},
	}
	writeJson(&conf, baseDir)
}

func mergeMaps(map1 map[string]string, map2 map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range map1 {
		merged[k] = v
	}
	for k, v := range map2 {
		merged[k] = v
	}
	return merged
}

func getJsonFilePath(baseDir string) string {
	return strings.Join([]string{baseDir, "conf/conf.json"}, "/")
}
