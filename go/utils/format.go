package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type GradleDep struct {
	group   string
	name    string
	version string
}

func Format(args []string) {
	cmd := args[1]
	if cmd == "grdep" {
		dep := parseDep(args[2])
		fmt.Printf("compile \"%s:%s:%s\"", dep.group, dep.name, dep.version)
	} else if cmd == "grlib" {
		dep := parseDep(args[2])
		varName := depNameToVarName(dep.name)
		fmt.Printf("%s: \"%s:%s:${versions.%s}\",", varName, dep.group, dep.name, varName)
		fmt.Print("\n")
		fmt.Printf("%s: \"%s\",", varName, dep.version)
	}
}

func parseDep(dep string) *GradleDep {
	if strings.Contains(dep, "group:") {
		r := regexp.MustCompile(".*?group:\\s*(\\S+).+?name:\\s*(\\S+).+?version:\\s*(\\S+)")
		parts := r.FindStringSubmatch(dep)
		return &GradleDep{
			group:   trim(parts[1]),
			name:    trim(parts[2]),
			version: trim(parts[3]),
		}
	} else {
		parts := strings.Split(dep, ":")
		if strings.Contains(dep, ":jar:") {
			return &GradleDep{
				group:   parts[0],
				name:    parts[1],
				version: parts[3],
			}
		}else{
			return &GradleDep{
				group:   parts[0],
				name:    parts[1],
				version: parts[2],
			}
		}
	}
	return nil
}

func depNameToVarName(str string) string {
	val := strings.ReplaceAll(str, ".", "_")
	val = strings.ReplaceAll(val, "-", "_")
	return val
}

func trim(str string) string {
	val := strings.ReplaceAll(str, " ", "")
	val = strings.ReplaceAll(val, "\"", "")
	val = strings.ReplaceAll(val, "'", "")
	val = strings.ReplaceAll(val, ",", "")
	return val
	//return strings.Trim(strings.Trim(str, "\""), "'")
}
