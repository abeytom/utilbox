package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func BearerToken(args []string) {
	baseDir := getBaseDir()
	cmd := args[1]
	conf := getConf(baseDir)
	if cmd == "add" {
		key := args[2]
		value := args[3]
		tokens := conf.Tokens
		tokens[key] = value
		writeJson(conf, baseDir)
	} else if cmd == "bh" {
		tokens := conf.Tokens
		key := args[2]
		if token, ok := tokens[key]; ok {
			fmt.Printf("Authorization: Bearer %s", token)
		} else {
			fmt.Fprintf(os.Stderr,"ERR:INVALID_TOKEN [%s]", key)
		}
	} else {
		tokens := conf.Tokens
		key := args[1]
		if token, ok := tokens[key]; ok {
			fmt.Print(token)
		} else {
			fmt.Fprintf(os.Stderr,"ERR:INVALID_TOKEN [%s]", key)
		}
	}
}

func getBaseDir() string {
	baseDir := os.Getenv("UTILBOX_PATH")
	if baseDir == "" {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		baseDir = filepath.Join(user.HomeDir, ".config/utilbox")
	}
	return baseDir
}

func GetKeyValue(key string) (string, error) {
	baseDir := getBaseDir()
	conf := getConf(baseDir)
	tokens := conf.Tokens
	if token, ok := tokens[key]; ok {
		return token, nil
	} else {
		return "", fmt.Errorf("ERR:INVALID_TOKEN [%s]", key)
	}
}
