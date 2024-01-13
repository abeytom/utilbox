package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/abeytom/utilbox/common"
	"gopkg.in/yaml.v2"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type awsParams struct {
	Ec2 ec2Params `yaml:"ec2"`
}

type ec2Params struct {
	InstanceType          string `yaml:"instanceType"`
	SshKeyPairName        string `yaml:"sshKeyPairName"`
	SecurityGroupId       string `yaml:"securityGroupId"`
	SubNetId              string `yaml:"subNetId"`
	AmiTypeTag            string `yaml:"amiTypeTag"`
	InstanceTypeTag       string `yaml:"instanceTypeTag"`
	InstanceNameTagPrefix string `yaml:"instanceNameTagPrefix"`
}

type awsDescribeInst struct {
	Reservations []awsReservation
}

type awsReservation struct {
	Instances []awsInstance
}

type awsInstance struct {
	ImageId         string
	InstanceId      string
	InstanceType    string
	PublicDnsName   string
	PublicIpAddress string
	Tags            []awsTags
	State           struct{ Name string }
	LaunchTime      string

	launchTimeMillis int64
	launchTimeHuman  string
}

type awsTags struct {
	Key   string
	Value string
}

type awsDescImg struct {
	Images []awsImage
}

type awsImage struct {
	Name               string
	ImageId            string
	State              string
	Description        string
	Tags               []awsTags
	VirtualizationType string
}

func AwsExec(args []string) {
	if len(args) == 0 {
		awsHelp()
	} else if args[0] == "ls" {
		params, _ := loadParams(args)
		awsList(args, params.Ec2)
	} else if args[0] == "stop" {
		awsInstanceCommand(args, "stop-instances", true)
	} else if args[0] == "start" {
		awsInstanceCommand(args, "start-instances", false)
	} else if args[0] == "terminate" {
		awsInstanceCommand(args, "terminate-instances", true)
	} else if args[0] == "launch" {
		params, argsMap := loadParams(args)
		awsLaunch(args, params.Ec2, argsMap)
	} else if args[0] == "create-ami" {
		params, argsMap := loadParams(args)
		createAmi(args, params.Ec2, argsMap)
	} else if args[0] == "ami" {
		if len(args) < 2 {
			awsHelp()
		} else if args[1] == "ls" {
			params, _ := loadParams(args)
			awsAmiList(args, params.Ec2)
		} else {
			awsHelp()
		}
	} else {
		awsHelp()
	}
}

func loadParams(args []string) (awsParams, map[string]string) {
	paramFile := os.Getenv("PARAM_FILE")
	if len(paramFile) == 0 {
		paramFile = filepath.Join(os.Getenv("HOME"), ".aws/params")
	}
	_, statErr := os.Stat(paramFile)
	if os.IsNotExist(statErr) {
		log.Fatalf("The params file doesnt exist at ~/.aws/params")
	}
	bytes, readErr := os.ReadFile(paramFile)
	if readErr != nil {
		log.Fatalf("The error while reading [%v]. The error is [%v]", paramFile, readErr)
	}
	var params awsParams
	parseErr := yaml.Unmarshal(bytes, &params)
	if parseErr != nil {
		log.Fatalf("Error parsing YAML: %v", parseErr)
	}
	ec2Inst := reflect.ValueOf(&params.Ec2).Elem()
	argsMap := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.Index(arg, "--ec2-") == 0 {
			varName := common.DelimToCamelCase(arg[6:], '-', true)
			field := ec2Inst.FieldByName(varName)
			if field.IsValid() && field.CanSet() {
				field.SetString(args[i+1])
				i++
			} else {
				argsMap[arg[6:]] = args[i+1]
				i++
			}
		}
	}
	return params, argsMap
}

func awsAmiList(args []string, ec2 ec2Params) {
	if len(ec2.AmiTypeTag) == 0 {
		log.Fatalf("The `amiTypeTag` must be set")
	}
	raw := len(args) > 2 && args[2] == "--raw"
	outStr, errStr, err := ExecuteCommand2("aws", "ec2", "describe-images", "--filters",
		"Name=tag:Type,Values="+ec2.AmiTypeTag)
	if err != nil {
		log.Fatalf("Error while listing the amis. The message is [%v]. The error is [%v]", errStr, err)
	}
	if raw {
		fmt.Println(outStr)
		return
	}
	var desc awsDescImg
	jsonErr := json.Unmarshal([]byte(outStr), &desc)
	if jsonErr != nil {
		log.Fatalf("Error while reading the response [%v]", err)
	}
	if len(desc.Images) == 0 {
		log.Fatalf("No AMIs found matching the filter")
	}
	var rows []DataRow
	for i, img := range desc.Images {
		row := DataRow{
			Cols: []interface{}{
				i + 1,
				img.ImageId,
				img.Name,
				img.State,
				img.VirtualizationType,
				awsTagToMap(img.Tags),
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
		Headers:      []string{"SL", "AMI", "NAME", "STATE", "VIRT", "TAGS"},
		GroupByCount: 0,
		Converted:    false,
	})
}

func createAmi(args []string, ec2 ec2Params, ec2ArgMap map[string]string) {
	if len(args) < 3 || len(strings.TrimSpace(args[1])) == 0 || len(strings.TrimSpace(args[2])) == 0 {
		log.Fatalf("The Instance ID and AMI Name should be set as args")
	}
	instanceId := strings.TrimSpace(args[1])
	amiName := strings.TrimSpace(args[2])
	typeTag := ensureValue(ec2.InstanceTypeTag, "instanceTypeTag")
	tagSpec := fmt.Sprintf("ResourceType=image,Tags=[{Key=Name,Value=%v},{Key=Type,Value=%v},{Key=owner,Value=abey}]",
		amiName, typeTag)
	ec2CmdArgs := []string{"ec2", "create-image",
		"--instance-id", instanceId,
		"--name", amiName,
		"--description", amiName,
		"--tag-specifications", tagSpec}
	for key, value := range ec2ArgMap {
		ec2CmdArgs = append(ec2CmdArgs, "--"+key, value)
	}
	cmdOut, cmdErr, err := ExecuteCommand2("aws", ec2CmdArgs...)
	if err != nil {
		log.Fatalf("Error while executing launch. The message is [%v]. The error is [%v]", cmdErr, err)
	}
	fmt.Println(cmdOut)
}

func awsLaunch(args []string, ec2 ec2Params, ec2ArgMap map[string]string) {
	if len(args) < 3 || len(strings.TrimSpace(args[1])) == 0 || len(strings.TrimSpace(args[2])) == 0 {
		log.Fatalf("The AMI ID and Suffix should be set as args")
	}
	amiId := strings.TrimSpace(args[1])
	suffix := strings.TrimSpace(args[2])

	instanceType := ensureValue(ec2.InstanceType, "instanceType")
	sshKeyPairName := ensureValue(ec2.SshKeyPairName, "sshKeyPairName")
	securityGroupId := ensureValue(ec2.SecurityGroupId, "securityGroupId")
	subNetId := ensureValue(ec2.SubNetId, "subNetId")
	typeTag := ensureValue(ec2.InstanceTypeTag, "instanceTypeTag")
	namePfx := ensureValue(ec2.InstanceNameTagPrefix, "instanceNameTagPrefix")
	ec2CmdArgs := []string{"ec2", "run-instances",
		"--image-id", amiId,
		"--instance-type", instanceType,
		"--key-name", sshKeyPairName,
		"--security-group-ids", securityGroupId,
		"--subnet-id", subNetId}
	inTags := ""
	for key, value := range ec2ArgMap {
		if key == "tag" {
			parts := strings.Split(value, "=")
			if len(parts) != 2 {
				continue
			}
			inTags += fmt.Sprintf(",{Key=%v,Value=%v}", parts[0], parts[1])
			continue
		}
		ec2CmdArgs = append(ec2CmdArgs, "--"+key, value)
	}
	tagSpec := fmt.Sprintf("ResourceType=instance,Tags=[{Key=Name,Value=%v%v},{Key=Type,Value=%v},{Key=owner,Value=abey},{Key=reason,Value=platform-test},{Key=expected-end-date,Value=12/12/2023}%v]",
		namePfx, suffix, typeTag, inTags)
	ec2CmdArgs = append(ec2CmdArgs, "--tag-specifications", tagSpec)

	cmdOut, cmdErr, err := ExecuteCommand2("aws", ec2CmdArgs...,
	)
	if err != nil {
		log.Fatalf("Error while executing launch. The message is [%v]. The error is [%v]", cmdErr, err)
	}
	fmt.Println(cmdOut)
}

func ensureValue(value string, key string) string {
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		log.Fatalf("The %v must be set", key)
	}
	return value
}

func awsInstanceCommand(args []string, cmd string, confirm bool) {
	if len(args) < 2 || len(strings.TrimSpace(args[1])) == 0 {
		log.Fatalf("The instance id must be set")
	}
	instanceId := strings.TrimSpace(args[1])
	if confirm {
		fmt.Printf("Do you want to %v %v [yes/no]: ", cmd, instanceId)
		reader2 := bufio.NewReader(os.Stdin)
		confirmStr, _ := reader2.ReadString('\n')
		if strings.TrimSpace(confirmStr) != "yes" {
			fmt.Printf("Skipping %v\n", cmd)
			return
		}
	}

	fmt.Printf("Executing `%v` on instance [%v]\n", cmd, instanceId)
	cmdOut, cmdErr, err := ExecuteCommand2("aws", "ec2", cmd, "--instance-ids", instanceId)
	if err != nil {
		log.Fatalf("Error while executing [%v]. The message is [%v]. The error is [%v]", cmd, cmdErr, err)
	}
	fmt.Println(cmdOut)
}

func awsList(args []string, param ec2Params) {
	if len(param.InstanceTypeTag) == 0 {
		log.Fatalf("The instanceTypeTag must be set")
	}
	raw := len(args) > 1 && args[1] == "--raw"
	outStr, errStr, err := ExecuteCommand2("aws", "ec2", "describe-instances", "--filters",
		"Name=tag:Type,Values="+param.InstanceTypeTag)
	if err != nil {
		log.Fatalf("Error while listing the instances. The message is [%v]. The error is [%v]", errStr, err)
	}
	if raw {
		fmt.Println(outStr)
		return
	}
	var desc awsDescribeInst
	jsonErr := json.Unmarshal([]byte(outStr), &desc)
	if jsonErr != nil {
		log.Fatalf("Error while reading the response [%v]", err)
	}
	if len(desc.Reservations) == 0 {
		log.Fatalf("No Instances found")
	}
	instances := getSortedInstances(desc)
	var rows []DataRow
	for i, inst := range instances {
		launchTime := inst.LaunchTime
		if len(inst.launchTimeHuman) > 0 {
			launchTime = launchTime + " (" + inst.launchTimeHuman + ")"
		}
		row := DataRow{
			Cols: []interface{}{
				i + 1,
				inst.InstanceId,
				inst.InstanceType,
				inst.State.Name,
				inst.PublicDnsName,
				inst.PublicIpAddress,
				awsTagToMap(inst.Tags),
				launchTime,
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
		Headers:      []string{"SL", "INSTANCE", "TYPE", "STATE", "HOSTNAME", "PUBLIC_IP", "TAGS", "LAUNCHED_AT"},
		GroupByCount: 0,
		Converted:    false,
	})
}

func getSortedInstances(desc awsDescribeInst) []awsInstance {
	layout := "2006-01-02T15:04:05-07:00"
	var instances []awsInstance
	for _, reservation := range desc.Reservations {
		for _, inst := range reservation.Instances {
			t, err := time.Parse(layout, inst.LaunchTime)
			if err == nil {
				inst.launchTimeHuman = timeSince(t)
				inst.launchTimeMillis = t.UnixMilli()
			} else {
				fmt.Println(err)
			}
			instances = append(instances, inst)
		}
	}
	//sort desc
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].launchTimeMillis > instances[j].launchTimeMillis
	})
	return instances
}

func awsTagToMap(tags []awsTags) map[string]interface{} {
	tagMap := make(map[string]interface{})
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}
	return tagMap
}

func timeSince(t time.Time) string {
	seconds := float64((time.Now().UnixMilli() - t.UnixMilli()) / 1000)
	interval := seconds / 31536000
	if interval > 1 {
		return strconv.Itoa(int(math.Round(interval))) + " years"
	}
	interval = seconds / 2592000
	if interval > 1 {
		return strconv.Itoa(int(math.Round(interval))) + " months"
	}
	interval = seconds / 86400
	if interval > 1 {
		return strconv.Itoa(int(math.Round(interval))) + " days"
	}
	interval = seconds / 3600
	if interval > 1 {
		return strconv.Itoa(int(math.Round(interval))) + " hours"
	}
	interval = seconds / 60
	if interval > 1 {
		return strconv.Itoa(int(math.Round(interval))) + " mins"
	}
	return strconv.Itoa(int(math.Round(interval))) + " secs"
}

func awsHelp() {
	fmt.Println("Available commands are:")
	fmt.Println("    awx ls [--raw]")
	fmt.Println("    aws ami ls [--raw]")
	fmt.Println("    awx launch <instanceId> <name suffix>")
	fmt.Println("    awx start <instanceId>")
	fmt.Println("    awx stop <instanceId>")
	fmt.Println("    awx terminate <instanceId>")
	fmt.Println("    awx create-ami <instanceId> <amiName>")
}
