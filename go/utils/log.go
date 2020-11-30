package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type LogLine struct {
	Timestamp  string `json:"@timestamp"`
	Thread     string `json:"thread_name"`
	Message    string `json:"message"`
	Level      string `json:"level"`
	Logger     string `json:"logger_name"`
	StackTrace string `json:"stack_trace"`
}

func JsonLog2Txt() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var logLine LogLine
		if err := json.Unmarshal(line, &logLine); err != nil {
			fmt.Println(string(line))
			continue
		}
		fmt.Printf("%s %s [%s] %s %s\n", logLine.Timestamp, logLine.Level, logLine.Thread, logLine.Logger, logLine.Message)
		if logLine.StackTrace != "" {
			fmt.Println(logLine.StackTrace)
		}
	}
}
