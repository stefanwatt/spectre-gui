package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type RipgrepResult = map[string][]RipgrepMatch

type RipgrepMatch struct {
	Path        string
	Row         string
	Col         string
	MatchedLine string
}

func (a *App) Search(search_term string, dir string, include string, exclude string) RipgrepResult {
	if search_term == "" {
		return RipgrepResult{}
	}
	lines := ripgrep(search_term, dir, include, exclude)
	matches := mapRipgrepMatch(lines)
	grouped := GroupByProperty(matches, func(match RipgrepMatch) string {
		return match.Path
	})
	return grouped
}

func map_glob_pattern(patterns string, exclude bool) []string {
	parts := strings.Split(patterns, ",")
	parts = Filter(parts, func(part string) bool { return part != "" })
	if len(parts) == 0 {
		return []string{}
	}
	x := MapArray(parts, func(pattern string) []string {
		var adapted_pattern string
		if exclude {
			adapted_pattern = fmt.Sprintf("!%s", pattern)
		} else {
			adapted_pattern = fmt.Sprintf("%s", pattern)
		}
		return []string{"--glob", adapted_pattern}
	})
	return Flatten(x)
}

func ripgrep(search_term string, dir string, include string, exclude string) []string {
	includeArgs := map_glob_pattern(include, false)
	excludeArgs := map_glob_pattern(exclude, true)
	args := append(includeArgs, excludeArgs...)
	args = append(args, []string{
		"-F",
		"--line-number",
		"--debug",
		"--column",
		"--no-heading",
		"--smart-case",
	}...)
	args = append(args, fmt.Sprintf("%s", search_term))
	args = append(args, fmt.Sprintf("%s", dir))
	rg_cmd := exec.Command("rg", args...)
	rg_cmd.Stderr = os.Stderr
	Log(rg_cmd.String())
	bytes, err := rg_cmd.Output()
	if err != nil {
		Log(err.Error())
		return []string{}
	}
	lines := strings.Split(string(bytes), "\n")
	lines = Filter(lines, func(line string) bool {
		return line != ""
	})
	return lines
}

func mapRipgrepMatch(lines []string) []RipgrepMatch {
	matches := MapArray(lines, func(line string) RipgrepMatch {
		regexpattern := `(.*):(.*):(.*):(.*)`
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
		return RipgrepMatch{
			filepath.Base(submatches[1]),
			submatches[2],
			submatches[3],
			matched_line,
		}
	})
	return matches
}

func GroupByProperty[T any, K comparable](items []T, getProperty func(T) K) map[K][]T {
	grouped := make(map[K][]T)

	for _, item := range items {
		key := getProperty(item)
		grouped[key] = append(grouped[key], item)
	}

	return grouped
}
