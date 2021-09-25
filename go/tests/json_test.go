package tests

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestKeys(t *testing.T) {
	podsJson := path.Join(getCurrentDir(t), "pods.json")

	cmd := fmt.Sprintf("cat %v | jp keys", podsJson)
	lines := execCmdGetLines(cmd)
	if len(lines) != 107 {
		log.Fatalf("Expected %v lines, found %v. The output is %v", 107, len(lines), lines)
	}
}

func TestKeyMultipleVals(t *testing.T) {
	podsJson := path.Join(getCurrentDir(t), "pods.json")

	cmd := fmt.Sprintf("cat %v | jp  keys[items.metadata.name,items.spec.containers.args]  out..table", podsJson)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines), 14)
	assertStringEquals(lines[0], "items.metadata.name            items.spec.containers.args    ")
	assertStringEquals(lines[1], "gateway-7b8c56d867-brgg7       nginx                         ")
	assertStringEquals(lines[2], "                               start                         ")

	cmd = fmt.Sprintf("cat %v | jp  keys[items.metadata.name,items.spec.containers.args]  out..csv", podsJson)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 8)
	assertStringEquals(lines[0], "items.metadata.name,items.spec.containers.args")
	assertStringEquals(lines[1], "gateway-7b8c56d867-brgg7,\"nginx,start\"")
}

func TestKeyBlankVals(t *testing.T) {
	podsJson := path.Join(getCurrentDir(t), "pods.json")

	cmd := fmt.Sprintf("cat %v | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..table head[name,ns,hostIp,podIp]", podsJson)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines), 8)
	assertStringEquals(lines[0], "name                           ns        hostIp          podIp           ")
	assertStringEquals(lines[1], "gateway-7b8c56d867-brgg7       sample    192.168.1.60    10.1.151.232    ")
	assertStringEquals(lines[6], "storefront-cd75b46c7-kl8jj     sample                                    ")

	cmd = fmt.Sprintf("cat %v | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..csv head[name,ns,hostIp,podIp]", podsJson)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 8)
	assertStringEquals(lines[0], "name,ns,hostIp,podIp")
	assertStringEquals(lines[1], "gateway-7b8c56d867-brgg7,sample,192.168.1.60,10.1.151.232")
	assertStringEquals(lines[6], "storefront-cd75b46c7-kl8jj,sample,,")
}

func TestCount(t *testing.T) {
	podsJson := path.Join(getCurrentDir(t), "pods.json")

	cmd := fmt.Sprintf("cat %v | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..table head[name,ns,hostIp,podIp] | csv row[1:] group[0]:count out..table tr..c0..split:-..merge:-..ncol[-1] sort[4]:desc", podsJson)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines), 7)
	assertStringEquals(lines[0], "name                     ns        hostIp          podIp           count    ")
	assertStringEquals(lines[1], "storefront-cd75b46c7     sample    192.168.1.60    10.1.151.233    3        ")
	assertStringEquals(lines[2], "                                                   10.1.151.234             ")

	cmd = fmt.Sprintf("cat %v | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..table head[name,ns,hostIp,podIp] | csv row[1:] group[0]:count out..csv tr..c0..split:-..merge:-..ncol[-1] sort[4]:desc", podsJson)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 5)
	assertStringEquals(lines[0], "name,ns,hostIp,podIp,count")
	assertStringEquals(lines[1], "storefront-cd75b46c7,sample,192.168.1.60,\"10.1.151.233,10.1.151.234\",3")
	assertStringEquals(lines[2], "gateway-7b8c56d867,sample,192.168.1.60,\"10.1.151.231,10.1.151.232\",2")
	assertStringEquals(lines[3], "data-store-844b74455c,sample,192.168.1.60,10.1.151.235,1")

	cmd = fmt.Sprintf("cat %v | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP] out..csv head[name,ns,hostIp,podIp] | csv split:csv row[1:] group[0]:count out..csv tr..c0..split:-..merge:-..ncol[-1] sort[4]:desc", podsJson)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 5)
	assertStringEquals(lines[0], "name,ns,hostIp,podIp,count")
	assertStringEquals(lines[1], "storefront-cd75b46c7,sample,192.168.1.60,\"10.1.151.233,10.1.151.234\",3")
	assertStringEquals(lines[2], "gateway-7b8c56d867,sample,192.168.1.60,\"10.1.151.231,10.1.151.232\",2")
	assertStringEquals(lines[3], "data-store-844b74455c,sample,192.168.1.60,10.1.151.235,1")
}

func assertStringEquals(actual string, expected string) {
	if actual != expected {
		err := fmt.Errorf("Expected String [%v] Actual [%v]", expected, actual)
		panic(err)
	}
}

func assertIntEquals(actual int, expected int) {
	if actual != expected {
		err := fmt.Errorf("Expected Value [%v] Actual [%v]", expected, actual)
		panic(err)
	}
}

func assertDeepEquals(actual interface{}, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		err := fmt.Errorf("Expected \n '%v' \n Actual \n '%v'\n", expected, actual)
		panic(err)
	}
}

func execCmdGetLines(cmd string) []string {
	out := execCmd(cmd)
	return strings.Split(string(out), "\n")
}

func execCmd(cmd string) []byte {
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatalf("Failed to execute command: %s", cmd)
	}
	return out
}

func getCurrentDir(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	//t.Logf("Current test filename: %s", filename)
	dir := path.Dir(filename)
	return dir
}
