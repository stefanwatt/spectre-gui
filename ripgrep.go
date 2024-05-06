package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type RipgrepResult = map[string][]RipgrepMatch

type RipgrepMatch struct {
	Path         string
	AbsolutePath string
	MatchedLine  string
	Row          int
	Col          int
}

func map_glob_pattern(patterns string, exclude bool) []string {
	parts := strings.Split(patterns, ",")
	parts = Filter(parts, func(part string) bool { return part != "" })
	if len(parts) == 0 {
		return []string{}
	}
	patterns2d := MapArray(parts, func(pattern string) []string {
		if exclude {
			pattern = fmt.Sprintf("!%s", pattern)
		}
		return []string{"--glob", pattern}
	})
	return Flatten(patterns2d)
}

func Ripgrep(search_term string, dir string, include string, exclude string) []RipgrepMatch {
	includeArgs := map_glob_pattern(include, false)
	excludeArgs := map_glob_pattern(exclude, true)
	args := append(includeArgs, excludeArgs...)
	args = append(args, []string{
		"-F",
		"--line-number",
		"--column",
		"--no-heading",
		"--smart-case",
	}...)
	args = append(args, search_term)
	args = append(args, dir)
	rg_cmd := exec.Command("rg", args...)
	rg_cmd.Stderr = os.Stderr
	bytes, err := rg_cmd.Output()
	if err != nil {
		Log(err.Error())
		return []RipgrepMatch{}
	}
	lines := strings.Split(string(bytes), "\n")
	lines = Filter(lines, func(line string) bool {
		return line != ""
	})
	return MapArray(lines, func(line string) RipgrepMatch {
		return map_ripgrep_match(line)
	})
}

func map_ripgrep_match(line string) RipgrepMatch {
	regexpattern := `(.*):(\d+):(\d+):(.*)`
	re := regexp.MustCompile(regexpattern)
	submatches := re.FindStringSubmatch(line)
	if len(submatches) != 5 {
		fmt.Println("matches ", submatches)
		fmt.Println("line ", line)
		panic("error parsing ripgrep output")
	}
	// length := float64(len(submatches[4]))
	// max_length := int(math.Min(25.0, length)) - 1
	matched_line := submatches[4]
	matched_line = strings.TrimSpace(matched_line)
	// if len(matched_line) > max_length && max_length > 0 {
	// 	matched_line = matched_line[:max_length]
	// }

	row, err := strconv.Atoi(submatches[2])
	if err != nil {
		panic(err)
	}
	col, err := strconv.Atoi(submatches[3])
	if err != nil {
		panic(err)
	}
	path := submatches[1]
	match := RipgrepMatch{
		filepath.Base(path),
		path,
		matched_line,
		row,
		col,
	}
	return match
}
