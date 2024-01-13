package utils

import (
	"regexp"
	"strings"
)

type TabularData struct {
	Headers []string            `json:"headers"`
	Rows    []map[string]string `json:"rows"`
}

func unscrape(str string) TabularData {
	var t TabularData
	if str == "" {
		return t
	}

	lines := strings.Split(str, "\n")
	if len(lines) < 2 {
		return t
	}

	pattern := regexp.MustCompile(`\s{3,}`)
	matches := make([]map[string]interface{}, 0)
	headers := make([]string, 0)
	wordStart := 0
	header := lines[0]

	for _, match := range pattern.FindAllStringIndex(header, -1) {
		sep := header[match[0]:match[1]]
		wordEnd := match[1] - len(sep)
		word := header[wordStart:wordEnd]
		matches = append(matches, map[string]interface{}{
			"word":  word,
			"start": wordStart,
			"end":   match[1],
		})
		headers = append(headers, word)
		wordStart = match[1]
	}

	word := header[wordStart:]
	matches = append(matches, map[string]interface{}{
		"word":  word,
		"start": wordStart,
		"end":   -1,
	})
	headers = append(headers, word)

	rows := make([]map[string]string, 0)
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		row := make(map[string]string)
		for _, match := range matches {
			end := match["end"].(int)
			if end == -1 {
				end = len(line)
			}
			value := strings.TrimSpace(line[match["start"].(int):end])
			if value == "" {
				value = "-"
			}
			row[match["word"].(string)] = value
		}
		rows = append(rows, row)
	}

	return TabularData{
		Headers: headers,
		Rows:    rows,
	}
}
