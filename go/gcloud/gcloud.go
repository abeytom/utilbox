package gcloud

import (
	"encoding/json"
	"fmt"
	"github.com/abeytom/cmdline-utils/k8"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type GcloudItem struct {
	CreateTime  string `json:"createTime"`
	Name        string `json:"name"`
	UpdateTime  string `json:"updateTime"`
	UpdateTime2 *time.Time
}

const DateFormat = "2006-01-02T15:04:05.999999Z"

func Execute(args []string) {
	//gcloud beta artifacts versions list --repository=maven-repo --location=us-west1 --package=com.inception.agent:agent-assembly -o json
	//gcloud beta artifacts packages list --repository=maven-repo --location=us-west1
	if len(args) == 0 || args[0] == "list" {
		packages := getSortedPackages()
		if packages == nil {
			return
		}
		for _, pkg := range packages {
			if len(args) > 0 && !filterMatches(pkg, args[1]) {
				continue
			}
			args := []string{"beta", "artifacts", "versions", "list",
				"--repository=maven-repo",
				"--location=us-west1",
				"--package=" + pkg.Name,
				"--format=json"}
			versionsStr, errOut, err := k8.ExecuteCommand("gcloud", args...)
			if err != nil || versionsStr == "" {
				fmt.Printf("[error] The version-list returned error for [%s]. The error is [%s]", pkg.Name, errOut)
				continue
			}
			versions := parseAndSort(versionsStr)
			latest := versions[0]
			paths := strings.Split(latest.Name, "/")
			item := fmt.Sprintf("%-60v", pkg.Name+":"+paths[len(paths)-1])
			diff := time.Now().Sub(*latest.UpdateTime2)
			fmt.Printf("%s%s\n", item, shortDur(diff))
		}
	}
}

func filterMatches(pkg *GcloudItem, filter string) bool {
	match, err := filepath.Match(filter, strings.Split(pkg.Name, ":")[1])
	return err == nil && match
}

func shortDur(d time.Duration) string {
	if d.Hours() >= 1 {
		hours := d.Hours()
		if hours < 24 {
			return fmt.Sprintf("(%v hours ago)", int(math.Round(hours)))
		}
		return fmt.Sprintf("(%v days ago)", int(math.Round(hours/24)))
	} else if d.Minutes() > 0 {
		return fmt.Sprintf("(%v minutes ago)", int(math.Round(d.Minutes())))
	} else {
		return "(a few seconds ago)"
	}
}

func getSortedPackages() []*GcloudItem {
	args := []string{"beta", "artifacts", "packages", "list", "--repository=maven-repo", "--location=us-west1", "--format=json"}
	pkgsJsonStr, errOut, err := k8.ExecuteCommand("gcloud", args...)
	if err != nil || pkgsJsonStr == "" {
		fmt.Printf("[error] The packages-list returned error. %s", errOut)
		return nil
	}
	return parseAndSort(pkgsJsonStr)
}

func parseAndSort(jsonStr string) []*GcloudItem {
	var pkgs []*GcloudItem
	json.Unmarshal([]byte(jsonStr), &pkgs)
	for _, pkg := range pkgs {
		updateTime, err := time.Parse(DateFormat, pkg.UpdateTime)
		if err != nil {
			fmt.Printf("The update time is not valid for package %s. The error is  [%s]\n", pkg, err)
			continue
		}
		pkg.UpdateTime2 = &updateTime
	}
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].UpdateTime2.After(*pkgs[j].UpdateTime2)
	})
	return pkgs
}
