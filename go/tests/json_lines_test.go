package tests

import (
	"fmt"
	"path"
	"testing"
)

func TestJsonLineKeys(t *testing.T) {
	fileStr := path.Join(getCurrentDir(t), "json.log")
	cmd := fmt.Sprintf("cat %v | jpl keys", fileStr)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines),6)
}

func TestJsonLineFilterWithOutKeys(t *testing.T) {
	fileStr := path.Join(getCurrentDir(t), "json.log")
	{
		cmd := fmt.Sprintf("cat %v | jpl 'filter..[fields] == \"\"'", fileStr)
		lines := execCmdGetLines(cmd)
		assertIntEquals(len(lines),2)
		assertStringEquals(lines[0],"{\"timestamp\":\"ts5\",\"message\":\"This is a message\"}")
		assertStringEquals(lines[1],"")
	}

	{
		cmd := fmt.Sprintf("cat %v | jpl 'filter..[fields] != \"\"'", fileStr)
		lines := execCmdGetLines(cmd)
		assertIntEquals(len(lines),5)
		assertStringEquals(lines[0],"{\"timestamp\":\"ts1\",\"message\":\"This is a message\",\"fields\":{\"pod\":\"pod1\",\"tag\":\"tag1\"}}")
		assertStringEquals(lines[4],"")
	}

	{
		cmd := fmt.Sprintf("cat %v | jpl 'filter..[fields.pod] == \"pod1\"'", fileStr)
		lines := execCmdGetLines(cmd)
		assertIntEquals(len(lines),3)
		assertStringEquals(lines[0],"{\"timestamp\":\"ts1\",\"message\":\"This is a message\",\"fields\":{\"pod\":\"pod1\",\"tag\":\"tag1\"}}")
		assertStringEquals(lines[2],"")
	}

	{
		cmd := fmt.Sprintf("cat %v | jpl 'filter..[fields.pod] != \"pod1\"'", fileStr)
		lines := execCmdGetLines(cmd)
		assertIntEquals(len(lines),4)
		assertStringEquals(lines[0],"{\"timestamp\":\"ts3\",\"message\":\"This is a message\",\"fields\":{\"pod\":\"pod2\",\"tag\":\"tag1\"}}")
		assertStringEquals(lines[3],"")
	}

	{
		cmd := fmt.Sprintf("cat %v | jpl 'filter..[timestamp] == \"ts4\"'", fileStr)
		lines := execCmdGetLines(cmd)
		assertIntEquals(len(lines),2)
		assertStringEquals(lines[0],"{\"timestamp\":\"ts4\",\"message\":\"This is a message\",\"fields\":{\"pod\":\"pod2\",\"tag\":\"tag1\"}}")
		assertStringEquals(lines[1],"")
	}
}

func TestJsonLineFilterWithKeys(t *testing.T) {
	fileStr := path.Join(getCurrentDir(t), "json.log")
	{
		cmd := fmt.Sprintf("cat %v | jpl keys[timestamp,fields] 'filter..[fields.pod] != \"pod1\"' out..table", fileStr)
		lines := execCmdGetLines(cmd)
		assertIntEquals(len(lines),8)
		assertStringEquals(lines[0],"ts3          pod: pod2    ")
		assertStringEquals(lines[7],"")
		for _, line := range lines {
			fmt.Println(line)
		}
	}
}