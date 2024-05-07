package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RipgrepResult = map[string][]RipgrepMatch

type RipgrepMatch struct {
	Id           string
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

func retryCommand(command string, args []string, retries int, delay time.Duration) (*string, error) {
	Log(fmt.Sprintf("Attempting to run %s", command))
	var err error
	for i := 0; i < retries; i++ {
		cmd := exec.Command(command, args...)
		cmd.Stderr = os.Stderr
		bytes, cmdErr := cmd.Output()

		if cmdErr == nil {
			output := string(bytes)
			Log(fmt.Sprintf("Successfully ran %s", command))
			Log(fmt.Sprintf("Output: %s", output))
			return &output, nil
		}
		fmt.Printf("Attempt %d failed: %s\n", i+1, cmdErr)
		fmt.Printf("Command: %s %s\n", command, strings.Join(args, " "))

		if i < retries-1 {
			fmt.Println("Waiting before retry...")
			time.Sleep(delay)
		}

		err = cmdErr
	}

	return nil, fmt.Errorf("command failed after %d attempts with error: %s", retries, err)
}

func Ripgrep(search_term string, dir string, include string, exclude string, flags []string) (*[]RipgrepMatch, error) {
	includeArgs := map_glob_pattern(include, false)
	excludeArgs := map_glob_pattern(exclude, true)
	args := append(includeArgs, excludeArgs...)
	args = append(args, []string{
		"--line-number",
		"--column",
		"--no-heading",
	}...)
	_, err := Find(flags, func(flag string) bool {
		return flag == "case_sensitive"
	})
	if err == nil {
		args = append(args, "--case-sensitive")
	} else {
		args = append(args, "--smart-case")
	}
	_, err = Find(flags, func(flag string) bool {
		return flag == "match_whole_word"
	})

	if err == nil {
		args = append(args, "--word-regexp")
	}
	_, err = Find(flags, func(flag string) bool {
		return flag == "regex"
	})
	if err != nil {
		args = append(args, "--fixed-strings")
	}
	args = append(args, search_term)
	args = append(args, dir)
	output, err := retryCommand("rg", args, 3, 1000)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(*output, "\n")
	lines = Filter(lines, func(line string) bool {
		return line != ""
	})
	matches := MapArray(lines, func(line string) RipgrepMatch {
		return map_ripgrep_match(line)
	})
	return &matches, nil
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
	uuid, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	match := RipgrepMatch{
		uuid.String(),
		filepath.Base(path),
		path,
		matched_line,
		row,
		col,
	}
	return match
}
