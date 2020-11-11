package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CsvExtract struct {
	Start int
	End   int
}

func HandleCsv(args []string) {
	//csv [delimiter]  [merge] [row_def] [col_def]

	//delimiter -> space(default) tab comma
	//merge -> default false
	//row_def -> row[0:] (default), row[1] row[2:3]
	//col_def -> col[0:] (default), col[1] col[2:3]

	// kc get pods | csv space row[1:] col[0]
	// kc get pods | csv space  col[0] -> gets all rows
	// kc get pods | csv space -> all lines will be merged
	// kc get pods | csv space merge ->
	//fmt.Printf("The args are %s\n", args)
	delim := " "
	merge := false
	mergeDelim := ", "
	filePath := args[0]
	rowExt := CsvExtract{0, -1}
	colExt := CsvExtract{0, -1}

	for _, arg := range args[1:] {
		if arg == "space" {
			delim = " "
		} else if arg == "tab" {
			delim = "\t"
		} else if arg == "comma" {
			delim = ","
		} else if strings.HasPrefix(arg, "merge") {
			merge = true
			if arg != "merge" {
				mergeDelim = extractMergeDelim(arg)
			}
		} else if strings.Index(arg, "row") == 0 {
			if arg != "row" {
				rowExt = extractCsvIndexArg(arg)
			}
		} else if strings.Index(arg, "col") == 0 {
			if arg != "col" {
				colExt = extractCsvIndexArg(arg)
			}
		} else if strings.Index(arg, "delim:") == 0 {
			delim = strings.Replace(arg, "delim:", "", 1)
		}
	}
	//fmt.Printf("delim %s, byLine %t, filePath=[%s], row %+v, col %+v \n",
	//	delim, merge, filePath, rowExt, colExt)

	file, err := os.Open(filePath)
	if err != nil {
		panic("Cannot read the file " + filePath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	idx := 0
	lines := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if isWithInBounds(rowExt, idx) {
			words := strings.Split(line, delim)
			extracted := extractCsv(words, colExt)
			lines = append(lines, strings.Join(extracted, ","))
		}
		idx = idx + 1
	}
	if merge {
		fmt.Printf("%s\n", strings.Join(lines, mergeDelim))
	} else {
		for _, line := range lines {
			fmt.Printf("%s\n", line)
		}
	}
}

func extractMergeDelim(arg string) string {
	delim := strings.Replace(arg, "merge:", "", 1)
	if delim == "comma" {
		return ", "
	}
	if delim == "space" {
		return " "
	}
	if delim == "tab" {
		return "\t"
	}
	return delim
}

func extractCsv(words []string, ext CsvExtract) []string {
	start := 0
	end := len(words) - 1
	if ext.Start != -1 {
		start = ext.Start
	}
	if ext.End != -1 {
		end = ext.End + 1
	}
	return words[start:end]
}

func isWithInBounds(ext CsvExtract, idx int) bool {
	upper := true
	lower := true
	if ext.Start != -1 {
		upper = idx >= ext.Start
	}
	if ext.End != -1 {
		if ext.Start != ext.End {
			lower = idx < ext.End
		} else {
			lower = idx == ext.End
		}
	}
	return upper && lower
}

func extractCsvIndexArg(arg string) CsvExtract {
	start := strings.Index(arg, "[")
	end := strings.Index(arg, "]")
	if start > 0 && end > start {
		str := string(([]rune(arg))[start+1 : end])
		if strings.Index(str, ":") != -1 {
			split := strings.Split(str, ":")
			if len(split) == 1 {
				return CsvExtract{strToInt(split[0], -1), -1}
			} else {
				return CsvExtract{strToInt(split[0], -1), strToInt(split[1], -1)}
			}
		} else {
			index := strToInt(str, -1)
			return CsvExtract{index, index}
		}
	} else {
		return CsvExtract{-1, -1}
	}
}
