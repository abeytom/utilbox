package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type CmdLine struct {
	Commands []string
	IArgs    map[string]string
	EArgs    map[string]string
}

func GxExec(args []string) {
	cmd := parseArgs(args)
	first := cmd.Commands[0]
	if first == "ls" {
		gxList(cmd)
	} else if first == "start" {
		gxInstanceCmd(cmd, "start")
	} else if first == "stop" {
		gxInstanceCmd(cmd, "stop")
	} else if first == "delete" {
		gxInstanceCmd(cmd, "delete")
	} else if first == "ssh" {
		gxSsh(cmd)
	} else if first == "scp" {
		gxScp(cmd)
	}
}

func gxInstanceCmd(cmd CmdLine, cmdName string) {
	instName := cmd.Commands[1]
	args := []string{"compute", "instances", cmdName, instName}
	args = appendExternalArgs(cmd, args)
	args = appendZoneArgs(cmd, instName, args)
	execGxCommand(cmd, args...)
}

func gxSsh(cmd CmdLine) {
	instName := cmd.Commands[1]
	args := []string{"compute", "ssh", "--tunnel-through-iap", instName}
	args = appendExternalArgs(cmd, args)
	args = appendZoneArgs(cmd, instName, args)
	fmt.Println("gcloud", strings.Join(args, " "))
}

func gxScp(cmd CmdLine) {
	args := []string{"compute", "scp", cmd.Commands[1], cmd.Commands[2]}
	args = appendExternalArgs(cmd, args)
	execGxCommand(cmd, args...)
}

func gxList(cmd CmdLine) {
	args := []string{"compute", "instances", "list"}
	if cmd.IArgs["all"] != "true" {
		args = append(args, "--filter", "labels.owner=atom")
	}
	command, _, _ := execGxCommand(cmd, args...)
	if len(command) <= 0 {
		return
	}
	data := unscrape(command)
	dir := getGxBaseDir()
	os.MkdirAll(dir, 0755)
	f := filepath.Join(dir, "gx-ls.json")
	marshal, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while marshalling json "+err.Error())
		return
	}
	err = os.WriteFile(f, marshal, 0655)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while writing json "+err.Error())
	}
}

func appendExternalArgs(cmd CmdLine, args []string) []string {
	for k, v := range cmd.EArgs {
		args = append(args, fmt.Sprintf("--%v=%v", k, v))
	}
	return args
}

func appendZoneArgs(cmd CmdLine, name string, args []string) []string {
	if !hasZoneArg(cmd) {
		zone := getGxInstanceZone(name)
		if len(zone) > 0 {
			args = append(args, "--zone", zone)
		}
	}
	return args
}

func hasZoneArg(cmd CmdLine) bool {
	for k, _ := range cmd.EArgs {
		if k == "zone" {
			return true
		}
	}
	return false
}

func getGxInstanceZone(name string) string {
	f := filepath.Join(getGxBaseDir(), "gx-ls.json")
	file, err := os.ReadFile(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while reading the file "+err.Error())
		return ""
	}
	var data TabularData
	err = json.Unmarshal(file, &data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while reading the file "+err.Error())
		return ""
	}
	zone := ""
	for _, row := range data.Rows {
		if row["NAME"] == name {
			if len(zone) == 0 {
				zone = row["ZONE"]
			} else {
				fmt.Fprintln(os.Stderr, "unable to uniquely identify the zone")
				return ""
			}
		}
	}
	return zone
}

func execGxCommand(cmd CmdLine, args ...string) (string, string, error) {
	for k, v := range cmd.EArgs {
		args = append(args, fmt.Sprintf("--%v=%v", k, v))
	}
	cmdOut, cmdErr, err := ExecuteCommand2("gcloud", args...)
	if err != nil {
		log.Fatalf("Error while executing [%v]. The message is [%v]. The error is [%v]", cmdOut, cmdErr, err)
	}
	if len(cmdErr) > 0 {
		fmt.Fprintln(os.Stderr, cmdErr)
	}
	fmt.Println(cmdOut)
	return cmdOut, cmdErr, err
}

func parseArgs(args []string) CmdLine {
	var cmds []string
	iArgsMap := make(map[string]string)
	eArgsMap := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		argType := getArgType(arg)
		if argType > 0 {
			key := arg[2:]
			value := ""
			if strings.Contains(key, "=") {
				parts := strings.Split(key, "=")
				key = parts[0]
				value = parts[1]
			} else if i < len(args)-1 {
				nextArgType := getArgType(args[i+1])
				if nextArgType == 0 {
					value = args[i+1]
					i++
				}
			}
			if len(value) == 0 {
				value = "true"
			}
			if argType == 1 {
				eArgsMap[key] = value
			} else if argType == 1 {
				iArgsMap[key] = value
			}
		} else {
			cmds = append(cmds, arg)
		}
	}
	return CmdLine{
		Commands: cmds,
		IArgs:    iArgsMap,
		EArgs:    eArgsMap,
	}
}

func getArgType(arg string) int {
	if strings.HasPrefix(arg, "--") {
		return 1
	}
	if strings.HasPrefix(arg, "..") {
		return 2
	}
	return 0
}

func getGxBaseDir() string {
	return filepath.Join(getBaseDir(), "gx")
}
