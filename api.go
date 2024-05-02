package main

import (
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type RipgrepMatch struct {
	Path        string
	Row         string
	Col         string
	MatchedLine string
}

func (a *App) Search(search_term string, dir string) []RipgrepMatch {
	return ripgrep(search_term, dir)
}

func ripgrep(search_term string, dir string) []RipgrepMatch {
	if search_term == "" || dir == "" {
		return []RipgrepMatch{}
	}
	rg_cmd := exec.Command("rg", "-F", "--line-number", "--column", "--no-heading", "--smart-case", search_term, dir)
	bytes, err := rg_cmd.Output()
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(bytes), "\n")
	lines = Filter(lines, func(line string) bool {
		return line != ""
	})
	matches := MapArray(lines, func(line string) RipgrepMatch {
		regexpattern := `(.*):(.*):(.*):(.*)`
		re := regexp.MustCompile(regexpattern)
		submatches := re.FindStringSubmatch(line)
		if len(submatches) != 5 {
			fmt.Println("matches ", submatches)
			fmt.Println("line ", line)
			panic("error parsing ripgrep output")
		}
		length := float64(len(submatches[4]))
		max_length := int(math.Min(25.0, length)) - 1
		matched_line := submatches[4]
		matched_line = matched_line[:max_length]
		return RipgrepMatch{
			filepath.Base(submatches[1]),
			submatches[2],
			submatches[3],
			matched_line,
		}
	})
	return matches

}
