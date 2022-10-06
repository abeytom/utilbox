package utils

import (
	"bufio"
	"fmt"
	"github.com/abeytom/utilbox/common"
	"os"
	"path/filepath"
	"strings"
)

type histOpts struct {
	dir *string
	cmd *string
}

// histx [-p | --pwd ] [-d | --dir /path/to/dir ] [-c | --cmd <cmd-to-filter>]
func ListHistory(args []string) {
	opts := parseHistArgs(args)
	fp := filepath.Join(os.Getenv("HOME"), ".zsh_history_detail")
	if !common.IsFile(fp) {
		fmt.Fprintf(os.Stderr, "History file not found at [%v]\n", fp)
		return
	}
	reader := newHistFileReader(fp)
	for {
		line := reader.nextLine()
		if len(line) == 0 {
			break
		}
		date, dir, cmd := splitHistLine(line)
		if len(cmd) == 0 {
			continue
		}
		if filterHistLine(date, dir, cmd, opts) {
			fmt.Println(date, dir, "--", cmd)
		}
	}
	defer reader.close()
}

func filterHistLine(date string, dir string, cmd string, opts histOpts) bool {
	if opts.dir != nil && dir != *opts.dir {
		return false
	}
	if opts.cmd != nil && !strings.HasPrefix(cmd, *opts.cmd) {
		return false
	}
	return true
}

func splitHistLine(line string) (string, string, string) {
	idx := 0
	var date string
	var dir string
	var command string
	for i, c := range line {
		if c == ',' {
			if idx == 0 {
				date = line[3:i]
				idx = i
			} else {
				dir = line[idx+1 : i]
				command = strings.TrimSpace(line[i+1:])
				break
			}
		}
	}
	return date, dir, command
}

func parseHistArgs(args []string) histOpts {
	var histData histOpts
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--pwd" || arg == "-p" {
			wd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			histData.dir = &wd
		} else if arg == "--dir" || arg == "-d" {
			argVal := args[i+1]
			histData.dir = &argVal
			i++
		} else if arg == "--cmd" || arg == "-c" {
			argVal := args[i+1]
			histData.cmd = &argVal
			i++
		}
	}
	return histData
}

type histFileReader struct {
	file    *os.File
	scanner *bufio.Scanner
	prev    string
}

func (h *histFileReader) close() {
	h.file.Close()
}

func (h *histFileReader) nextLine() string {
	tmp := h.prev
	if len(tmp) > 0 {
		tmp += "\n"
	}
	scanned := false
	for h.scanner.Scan() {
		scanned = true
		nline := h.scanner.Text()
		size := len(nline)
		if size > 0 {
			if nline[size-1] == '\\' {
				nline = nline[0 : size-1]
			}
		}
		if strings.HasPrefix(nline, "^^^") {
			if len(tmp) == 0 {
				tmp = nline + "\n"
				continue
			} else {
				h.prev = nline
				break
			}
		}
		tmp += nline + "\n"
	}
	if !scanned {
		h.prev = ""
	}
	return strings.TrimSpace(tmp)
}

func newHistFileReader(fp string) *histFileReader {
	file, err := os.Open(fp)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	return &histFileReader{
		file:    file,
		scanner: scanner,
	}
}
