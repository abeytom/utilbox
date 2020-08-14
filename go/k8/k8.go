package k8

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

/**
export k8ns=namespace
kc pod, kc log, kc ssh, kc secret

kc <other args for kubectl except namespace> | kc get pods | kc get services
kc logs pod-name* | kc logs pod-name* containerName | kc logs `kc pod admin-server*` containerName
kc ssh pod-name* | kc ssh pod-name* --container <container> | kc ssh `kc pod admin-server*` --container <container>
kc pod pod-name* | kc pod pod-name* <match-index-integer> | will sort matches alphabetically
kc get secrets | list secrets
kc secret secret-name* | (kc secret secret-name* | pbc)
*/
func Execute(allArgs []string) {
	//fmt.Printf("The args are %s\n",allArgs)
	ns := os.Getenv("k8ns")
	if ns == "" {
		ns = "default"
	}
	args := allArgs[1:]
	cmdFile := args[0]

	os.Setenv("k8exec", "")

	if args[1] == "log" || args[1] == "logs" {
		name, err := getPodByName(ns, args[2], -1, true)
		if err != nil || name == "" {
			return
		}
		restArgs := args[3:]
		args := append([]string{"kubectl", "-n", ns, "logs", name}, restArgs...)
		ioutil.WriteFile(cmdFile, []byte(strings.Join(args, " ")), 0644)
	} else if args[1] == "ssh" {
		name, err := getPodByName(ns, args[2], -1, true)
		if err != nil || name == "" {
			return
		}
		restArgs := args[3:]
		if len(restArgs) == 0 {
			restArgs = []string{"bash"}
		}
		args := append([]string{"kubectl", "-n", ns, "exec", "-it", name}, restArgs...)
		ioutil.WriteFile(cmdFile, []byte(strings.Join(args, " ")), 0644)
	} else if args[1] == "pod" {
		var matchIndex = -1
		if len(args) > 3 {
			index, err := strconv.Atoi(args[3])
			if err != nil {
				matchIndex = -1
			} else {
				matchIndex = index
			}
		}
		name, err := getPodByName(ns, args[2], matchIndex, false)
		if err != nil || name == "" {
			return
		}
		args := []string{"echo", name}
		ioutil.WriteFile(cmdFile, []byte(strings.Join(args, " ")), 0644)
	} else if args[1] == "secret" {
		var matchIndex = -1
		if len(args) > 3 {
			index, err := strconv.Atoi(args[3])
			if err != nil {
				matchIndex = -1
			} else {
				matchIndex = index - 1
			}
		}
		name, err := getSecretByName(ns, args[2], matchIndex)
		if err != nil || name == "" {
			return
		}
		str, err := getSecretStr(name, ns)
		if str != "" {
			args2 := []string{"echo", str}
			ioutil.WriteFile(cmdFile, []byte(strings.Join(args2, " ")), 0644)
		}
	} else if args[1] == "events" || args[1] == "event" {
		args := append([]string{"kubectl", "-n", ns, "get", "events", "--sort-by='{.lastTimestamp}'"})
		ioutil.WriteFile(cmdFile, []byte(strings.Join(args, " ")), 0644)
	} else {
		restArgs := args[1:]
		args := append([]string{"kubectl", "-n", ns}, restArgs...)
		//fmt.Printf("Running the command %s\n", args)
		ioutil.WriteFile(cmdFile, []byte(strings.Join(args, " ")), 0644)
	}
}

func getSecretStr(name string, namespace string) (string, error) {
	args := []string{"-n", namespace, "describe", "secret", name}
	secretStr, errOut, err := ExecuteCommand("kubectl", args...)
	if err != nil || secretStr == "" {
		fmt.Printf("[error] The get pods returned error. %s", errOut)
		return "", nil
	}
	secretListStr := string(secretStr[:])
	scanner := bufio.NewScanner(strings.NewReader(secretListStr))
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Index(text, "token:") != -1 {
			replaced := strings.Replace(text, "token:", "", 1)
			return strings.Trim(replaced, " "), nil
		}
	}
	fmt.Print("[error] Couldnt get the token from the secret", errOut)
	return "", nil
}

func getSecretByName(namespace string, nameMatch string, matchIndex int) (string, error) {
	args := []string{"-n", namespace, "get", "secrets"}
	secretStr, errOut, err := ExecuteCommand("kubectl", args...)
	if err != nil || secretStr == "" {
		fmt.Printf("[error] The get pods returned error. %s", errOut)
		return "", nil
	}
	secretListStr := string(secretStr[:])
	scanner := bufio.NewScanner(strings.NewReader(secretListStr))
	matchedSecrets := []string{}
	for scanner.Scan() {
		text := scanner.Text()
		podName := strings.Split(text, " ")[0]
		match, err := filepath.Match(nameMatch, podName)
		if err != nil {
			fmt.Println(err)
		}
		if match {
			matchedSecrets = append(matchedSecrets, podName)
		}
	}
	if matchIndex >= len(matchedSecrets) {
		fmt.Printf("[error] Invalid IEFilter Index")
		return "", nil
	}
	if len(matchedSecrets) == 0 {
		fmt.Printf("[error] Cannot find a secret in this namespace %s with name prefix %s\n", namespace, nameMatch)
	} else if len(matchedSecrets) > 1 {
		if matchIndex >= 0 {
			return matchedSecrets[matchIndex], nil
		}
		fmt.Printf("[error] Multiple Matches found for namespace %s with name prefix %s\n", namespace, nameMatch)
		sort.Strings(matchedSecrets)
		for i, name := range matchedSecrets {
			fmt.Printf("%d. %s\n", i+1, name)
		}
	} else {
		return matchedSecrets[0], nil
	}
	return "", nil
}

func getPodByName(namespace string, podMatch string, matchIndex int, runningOnly bool) (string, error) {
	args := []string{"-n", namespace, "get", "pods"}
	podsStr, errOut, err := ExecuteCommand("kubectl", args...)
	if err != nil || podsStr == "" {
		fmt.Printf("[error] The get pods returned error. %s", errOut)
		return "", nil
	}

	podListStr := string(podsStr[:])
	scanner := bufio.NewScanner(strings.NewReader(podListStr))
	matchedPods := []string{}
	header := true
	for scanner.Scan() {
		if header {
			header = false
			continue
		}
		text := scanner.Text()
		vals := strings.Fields(text)
		podName := vals[0]
		if runningOnly {
			if vals[2] != "Running" {
				continue
			}
		}
		match, err := filepath.Match(podMatch, podName)
		if err != nil {
			fmt.Println(err)
		}
		if match {
			matchedPods = append(matchedPods, podName)
		}
	}
	if matchIndex >= len(matchedPods) {
		fmt.Printf("[error] Invalid Match Index\n")
		return "", nil
	}
	if len(matchedPods) == 0 {
		fmt.Printf("[error] Cannot find a pod in this namespace %s with name prefix %s\n", namespace, podMatch)
	} else if len(matchedPods) > 1 {
		if matchIndex >= 0 {
			return matchedPods[matchIndex], nil
		}
		fmt.Printf("[error] Multiple Matches found for namespace %s with name prefix %s\n", namespace, podMatch)
		sort.Strings(matchedPods)
		for i, name := range matchedPods {
			fmt.Printf("%d. %s\n", i, name)
		}
	} else {
		return matchedPods[0], nil
	}
	return "", nil
}

func ExecuteCommand(cmdName string, args ...string) (string, string, error) {
	//fmt.Printf("[info] Running the command [%s %s]\n", cmdName, args)
	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("[error] " + fmt.Sprint(err) + ": " + stderr.String())
		return "", stderr.String(), err
	}
	return out.String(), "", nil
}
