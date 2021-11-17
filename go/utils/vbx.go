package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/abeytom/utilbox/common"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func VbExec(args []string) {
	if len(args) == 0 {
		vbHelp()
	} else if args[0] == "clone" {
		VbClone(args)
	} else if args[0] == "start" {
		VbStart(args)
	} else if args[0] == "stop" {
		VbStop(args)
	} else if args[0] == "save" {
		VbStop(args)
	} else if args[0] == "ls" {
		VbStatus(args)
	} else if args[0] == "rm" {
		VbRemove(args)
	} else {
		vbHelp()
	}
}

func vbHelp() {
	fmt.Println("Available commands are:")
	fmt.Println("    vbx ls")
	fmt.Println("    vbx clone")
	fmt.Println("    vbx start")
	fmt.Println("    vbx stop")
	fmt.Println("    vbx save")
	fmt.Println("    vbx rm")
}

func VbRemove(args []string) {
	vms := getAllVms()
	if len(vms) == 0 {
		fmt.Println("No VMs are present")
		return
	}
	for i, vm := range vms {
		fmt.Printf("%v. %v\n", i+1, vm)
	}
	fmt.Print("Enter a numeric choice: ")
	reader := bufio.NewReader(os.Stdin)
	choiceStr, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(strings.TrimSpace(choiceStr))
	if err != nil || choice > len(vms) {
		log.Fatalf("Invalid Choice %v", err)
	}
	chosen := vms[choice-1]
	fmt.Printf("Do you want to delete the VM '%v' [yes/no]: ", chosen)
	reader2 := bufio.NewReader(os.Stdin)
	confirmStr, _ := reader2.ReadString('\n')
	if strings.TrimSpace(confirmStr) == "yes" {
		startOut, startErrOut, startErr := ExecuteCommand2("VBoxManage", "unregistervm",
			"--delete", chosen)
		if startErr != nil {
			log.Fatalf("%v. %v", startErrOut, startErr)
		}
		fmt.Println(startOut)
		fmt.Println("Deleted the VM ", chosen)
	} else {
		fmt.Println("Skipping delete")
	}
}

func VbStatus(args []string) {
	stOut, stErrOut, stErr := ExecuteCommand2("VBoxManage", "list",
		"vms", "--long")
	if stErr != nil {
		fmt.Printf("[ERROR] %v. %v\n", stErrOut, stErr)
	}
	lines := strings.Split(stOut, "\n")
	var list []map[string]string
	var kvMap map[string]string
	for _, line := range lines {
		if strings.Index(line, "Name:") == 0 {
			kvMap = make(map[string]string)
			list = append(list, kvMap)
		}
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		kvMap[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	var rows []DataRow
	for i, kv := range list {
		row := DataRow{
			Cols: []interface{}{
				i + 1,
				kv["Name"],
				kv["Memory size"],
				kv["Number of CPUs"],
				kv["State"],
			},
		}
		rows = append(rows, row)
	}
	csvFmt := &CsvFormat{
		ColExt:      &common.IntRange{},
		RowExt:      &common.IntRange{},
		Split:       "space+",
		Merge:       " ",
		LMerge:      ",",
		Wrap:        "",
		IsLMerge:    false,
		NoHeaderOut: false,
		OutputDef:   &OutputDef{Type: "table"},
	}
	processOutput(csvFmt, &DataRows{
		DataRows:     rows,
		Headers:      []string{"SL", "NAME", "MEMORY", "CPU", "STATE"},
		GroupByCount: 0,
		Converted:    false,
	})
}

func VbStop(args []string) {
	vms := getRunningVms()
	if len(vms) == 0 {
		fmt.Println("No VMs are running")
		return
	}
	for _, vm := range vms {
		fmt.Printf("Saving State %v \n", vm)
		stopOut, stopErrOut, stopErr := ExecuteCommand2("VBoxManage", "controlvm",
			vm, "savestate")
		if stopErr != nil {
			fmt.Printf("[ERROR] %v. %v\n", stopErrOut, stopErr)
		}
		fmt.Println(stopOut)
	}
}

func VbStart(args []string) {
	runningVms := getRunningVms()
	if len(runningVms) > 0 {
		//make it a confirmation
		log.Fatalf("there are Vms running %v\n", runningVms)
	}
	vms := getAllVms()
	for i, vm := range vms {
		fmt.Printf("%v. %v\n", i+1, vm)
	}
	fmt.Print("Enter a numeric choice: ")
	reader := bufio.NewReader(os.Stdin)
	choiceStr, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(strings.TrimSpace(choiceStr))
	if err != nil || choice > len(vms) {
		log.Fatalf("Invalid Choice %v", err)
	}
	chosen := vms[choice-1]
	startOut, startErrOut, startErr := ExecuteCommand2("VBoxManage", "startvm", chosen)
	if startErr != nil {
		log.Fatalf("%v. %v", startErrOut, startErr)
	}
	fmt.Println(startOut)
}

func VbClone(args []string) {
	nSuffix := "cloned"
	if len(args) > 1 {
		nSuffix = args[1]
	}
	runningVms := getRunningVms()
	if len(runningVms) > 0 {
		//make it a confirmation
		log.Fatalf("there are Vms running %v\n", runningVms)
	}

	var baseVms []string
	var nonBaseVms []string
	for _, vm := range getAllVms() {
		if strings.Index(vm, ".BASE") != -1 {
			baseVms = append(baseVms, vm)
		} else {
			nonBaseVms = append(nonBaseVms, vm)
		}
	}
	baseCount := len(baseVms)
	vms := append(baseVms, nonBaseVms...)
	for i, vm := range vms {
		if i == baseCount {
			fmt.Println("---- Non Base Vms ----")
		}
		fmt.Printf("%v. %v\n", i+1, vm)
	}

	fmt.Print("Enter a numeric choice: ")
	reader := bufio.NewReader(os.Stdin)
	choiceStr, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(strings.TrimSpace(choiceStr))
	if err != nil || choice > len(vms) {
		log.Fatalf("Invalid Choice %v", err)
	}
	chosen := vms[choice-1]
	cloneName := strings.Replace(chosen, ".BASE", "", -1) + "." + nSuffix

	fmt.Printf("The cloning the VM %v as %v\n", chosen, cloneName)

	cloneOut, cloneErrOut, cloneErr := ExecuteCommand2("VBoxManage", "clonevm", chosen,
		fmt.Sprintf("--name=%v", cloneName),
		"--register", "--mode=machine", "--options=KeepAllMACs")
	if cloneErr != nil {
		log.Fatalf("%v. %v", cloneErrOut, cloneErr)
	}
	fmt.Println(cloneOut)
	startOut, startErrOut, startErr := ExecuteCommand2("VBoxManage", "startvm", cloneName)
	if startErr != nil {
		log.Fatalf("%v. %v", startErrOut, startErr)
	}
	fmt.Println(startOut)
}

func getBaseVms(vms []string) []string {
	var names []string
	for _, vm := range vms {
		if strings.Index(vm, ".BASE") != -1 {
			names = append(names, vm)
		}
	}
	return names
}

func getAllVms() []string {
	command, errOut, err := ExecuteCommand2("VBoxManage", "list", "vms")
	if err != nil {
		log.Fatalf("%v. %v", errOut, err)
	}
	var rx = regexp.MustCompile(`.*"(.+?)".*`)
	lines := strings.Split(command, "\n")
	var vms []string
	for _, line := range lines {
		submatch := rx.FindAllStringSubmatch(line, 1)
		if len(submatch) != 1 {
			continue
		}
		vms = append(vms, submatch[0][1])

	}
	return vms
}

func getRunningVms() []string {
	command, errOut, err := ExecuteCommand2("VBoxManage", "list", "runningvms")
	if err != nil {
		log.Fatalf("%v. %v", errOut, err)
	}
	var rx = regexp.MustCompile(`.*"(.+?)".*`)
	lines := strings.Split(command, "\n")
	var vms []string
	for _, line := range lines {
		submatch := rx.FindAllStringSubmatch(line, 1)
		if len(submatch) != 1 {
			continue
		}
		vms = append(vms, submatch[0][1])

	}
	return vms
}

func ExecuteCommand2(cmdName string, args ...string) (string, string, error) {
	//fmt.Printf("[info] Running the command [%s %s]\n", cmdName, args)
	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("[error] %v: %v", fmt.Sprint(err), stderr.String())
		return "", stderr.String(), err
	}
	return out.String(), "", nil
}
