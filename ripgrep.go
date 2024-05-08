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
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type RipgrepResult = map[string][]RipgrepMatch

type RipgrepMatch struct {
	Id              string
	Path            string
	AbsolutePath    string
	MatchedLine     string
	TextBeforeMatch string
	TextAfterMatch  string
	MatchedText     string
	ReplacementText string
	Row             int
	Col             int
}

func Ripgrep(
	search_term string,
	dir string,
	include string,
	exclude string,
	flags []string,
	replace_term string,
	preserve_case bool,
) (*[]RipgrepMatch, error) {
	includeArgs := map_glob_pattern(include, false)
	excludeArgs := map_glob_pattern(exclude, true)
	args := append(includeArgs, excludeArgs...)
	args = append(args, []string{
		"--line-number",
		"--column",
		"--no-heading",
		"--vimgrep",
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
		return map_ripgrep_match(line, search_term, replace_term, preserve_case)
	})
	return &matches, nil
}

func map_ripgrep_match(line string, search_term string, replace_term string, preserve_case bool) RipgrepMatch {
	regexpattern := `^(.*):(\d+):(\d+):(.*)$`
	re := regexp.MustCompile(regexpattern)
	submatches := re.FindStringSubmatch(line)
	if len(submatches) != 5 {
		fmt.Println("Error parsing line:", line)
		panic("incorrect format from ripgrep output")
	}
	matched_line := strings.TrimSpace(submatches[4])
	re_match := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(search_term))
	matched_text := re_match.FindString(matched_line)

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
	replacement_text := replace_term
	if preserve_case {
		replacement_text = map_replacement_text_preserve_case(matched_text, replace_term)
	}
	before, after := map_before_and_after(matched_line, matched_text)
	match := RipgrepMatch{
		uuid.String(),
		filepath.Base(path),
		path,
		matched_line,
		before,
		after,
		matched_text,
		replacement_text,
		row,
		col,
	}
	return match
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

func map_replacement_text_preserve_case(matched_text string, replace_term string) string {
	Log(fmt.Sprintf("mapping replacement text: \nmatched_text:%s\nreplace_term:%s", matched_text, replace_term))
	titleCaser := cases.Title(language.English)
	// ALL UPPERCASE
	if matched_text == "" {
		return replace_term
	}
	if matched_text == strings.ToUpper(matched_text) {
		return strings.ToUpper(replace_term)
	}
	// FIRST LETTER UPPER
	if len(matched_text) > 0 && unicode.IsUpper(rune(matched_text[0])) && matched_text[1:] == strings.ToLower(matched_text[1:]) {
		return titleCaser.String(replace_term)
	}
	// DEFAULT
	return replace_term
}

func map_before_and_after(matched_line, matched_text string) (string, string) {
	// Find the index of the first occurrence of matched_text in matched_line
	index := strings.Index(strings.ToLower(matched_line), strings.ToLower(matched_text))
	if index == -1 {
		// If the text is not found, return empty strings
		return "", ""
	}

	// Calculate the end index of the matched text
	end := index + len(matched_text)

	// Extract the text before and after the matched text
	textBeforeMatch := ""
	if index > 0 {
		textBeforeMatch = matched_line[:index]
	}
	textAfterMatch := ""
	if end < len(matched_line) {
		textAfterMatch = matched_line[end:]
	}

	return textBeforeMatch, textAfterMatch
}
