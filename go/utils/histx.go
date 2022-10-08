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
	dir     *string
	cmd     *string
	sid     *string
	suniq   bool
	uniq    bool
	cmdOnly bool
}

type histLine struct {
	date string
	dir  string
	sid  string
	cmd  string
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
		hist := splitHistLine(line)
		if len(hist.cmd) == 0 {
			continue
		}
		if filterHistLine(hist, opts) {
			if opts.cmdOnly {
				fmt.Println(hist.cmd)
			} else {
				fmt.Println(hist.date, hist.dir, hist.sid, "||", hist.cmd)
			}
		}
	}
	defer reader.close()
}

type filterState struct {
	prev  string
	lines map[string]bool
}

var state filterState

func filterHistLine(hist histLine, opts histOpts) bool {
	if opts.dir != nil && hist.dir != *opts.dir {
		return false
	}
	if opts.cmd != nil && !strings.HasPrefix(hist.cmd, *opts.cmd) {
		return false
	}
	if opts.sid != nil && hist.sid != *opts.sid {
		return false
	}
	if opts.uniq {
		if _, ok := state.lines[hist.cmd]; ok {
			return false
		}
		state.lines[hist.cmd] = true
	} else if opts.suniq {
		if hist.cmd == state.prev {
			return false
		}
		state.prev = hist.cmd
	}
	return true
}

func splitHistLine(line string) histLine {
	idx := 0
	count := 0
	var date string
	var dir string
	var sid string
	var command string
	for i, c := range line {
		if c == ',' {
			if count == 0 {
				date = line[3:i]
				idx = i
				count++
			} else if count == 1 {
				dir = line[idx+1 : i]
				idx = i
				count++
			} else {
				sid = line[idx+1 : i]
				command = strings.TrimSpace(line[i+1:])
				break
			}
		}
	}
	return histLine{
		date: date,
		dir:  dir,
		sid:  sid,
		cmd:  command,
	}
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
		} else if arg == "--session" || arg == "-s" {
			argVal := args[i+1]
			histData.sid = &argVal
			i++
		} else if arg == "--uniq" {
			histData.uniq = true
			state.lines = make(map[string]bool)
		} else if arg == "--suniq" {
			histData.suniq = true
		} else if arg == "--co" {
			histData.cmdOnly = true
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
