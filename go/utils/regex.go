package utils

import (
	"encoding/csv"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type regexInput struct {
	regex      []*regexp.Regexp
	cols       int
	groupCount int
}

func RegexExtract(args []string) {
	input := processArgs(args)
	regexList := input.regex
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()
	readStdIn2(func(line []byte) {
		for _, regex := range regexList {
			matches := regex.FindSubmatch(line)
			size := len(matches)
			if size > 1 {
				row := make([]string, input.cols)
				i := 1
				for ; i < size; i++ {
					row[i-1] = string(matches[i])
				}
				for ; i < input.cols; i++ {
					row[i-1] = ""
				}
				writer.Write(row)
				break
			}
		}
	})
}

func processArgs(args []string) regexInput {
	input := regexInput{}
	for _, arg := range args {
		if strings.Index(arg, "cols:") == 0 {
			cols, err := strconv.Atoi(arg[5:])
			if err != nil {
				log.Fatalf("The cols val [%v] should be an integer. The error is %v",
					arg[5:], err)
			}
			input.cols = cols
		} else {
			regex := regexp.MustCompile(arg)
			if regex.NumSubexp() < 1 {
				log.Fatalf("The regex must contain atleast one group enclosed in (..)")
			}
			input.regex = append(input.regex, regex)
			input.groupCount = MaxInt(input.groupCount, regex.NumSubexp())
		}
	}
	if len(input.regex) == 0 {
		log.Fatal("The regex expression must be set")
	}
	if input.cols > 0 && input.cols < input.groupCount {
		log.Fatal("The cols value must be greater than or equals the number of regex groups")
	}
	if input.cols == 0 {
		input.cols = input.groupCount
	}
	return input
}

func MaxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}
