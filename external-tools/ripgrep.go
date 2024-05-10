package externaltools

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	utils "spectre-gui/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Ripgrep(
	search_term string,
	dir string,
	include string,
	exclude string,
	flags []string,
	replace_term string,
	preserve_case bool,
) ([]string, error) {
	includeArgs := map_glob_pattern(include, false)
	excludeArgs := map_glob_pattern(exclude, true)
	args := append(includeArgs, excludeArgs...)
	args = append(args, []string{
		"--line-number",
		"--column",
		"--no-heading",
		"--vimgrep",
		"--only-matching",
	}...)
	_, err := utils.Find(flags, func(flag string) bool {
		return flag == "case_sensitive"
	})
	if err == nil {
		args = append(args, "--case-sensitive")
	} else {
		args = append(args, "--smart-case")
	}
	_, err = utils.Find(flags, func(flag string) bool {
		return flag == "match_whole_word"
	})

	if err == nil {
		args = append(args, "--word-regexp")
	}
	_, err = utils.Find(flags, func(flag string) bool {
		return flag == "regex"
	})
	if err != nil {
		args = append(args, "--fixed-strings")
	}
	args = append(args, search_term)
	args = append(args, dir)
	output, err := retryCommand("rg", args, 3, 1000)
	if err != nil {
		return []string{}, err
	}

	rg_output_lines := strings.Split(*output, "\n")
	rg_output_lines = utils.Filter(rg_output_lines, func(line string) bool {
		return line != ""
	})
	return rg_output_lines, nil
}

type RipgrepInfo struct {
	Path        string
	MatchedText string
	Row         int
	Col         int
}

func MapRipgrepInfo(output string) RipgrepInfo {
	regexpattern := `^(.*):(\d+):(\d+):(.*)$`
	re := regexp.MustCompile(regexpattern)
	submatches := re.FindStringSubmatch(output)

	if len(submatches) != 5 {
		fmt.Println("Error parsing line:", output)
		panic("incorrect format from ripgrep output")
	}
	path := submatches[1]
	matched_text := submatches[4]
	row, err := strconv.Atoi(submatches[2])
	if err != nil {
		panic(err)
	}

	col, err := strconv.Atoi(submatches[3])
	if err != nil {
		panic(err)
	}
	return RipgrepInfo{path, matched_text, row, col}
}

func map_glob_pattern(patterns string, exclude bool) []string {
	parts := strings.Split(patterns, ",")
	parts = utils.Filter(parts, func(part string) bool { return part != "" })
	if len(parts) == 0 {
		return []string{}
	}
	patterns2d := utils.MapArray(parts, func(pattern string) []string {
		if exclude {
			pattern = fmt.Sprintf("!%s", pattern)
		}
		return []string{"--glob", pattern}
	})
	return utils.Flatten(patterns2d)
}

func retryCommand(command string, args []string, retries int, delay time.Duration) (*string, error) {
	utils.Log(fmt.Sprintf("Attempting to run %s", command))
	var err error
	for i := 0; i < retries; i++ {
		cmd := exec.Command(command, args...)
		cmd.Stderr = os.Stderr
		bytes, cmdErr := cmd.Output()

		if cmdErr == nil {
			output := string(bytes)
			utils.Log(fmt.Sprintf("Successfully ran %s", command))
			utils.Log(fmt.Sprintf("Output: %s", output))
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
	utils.Log(fmt.Sprintf("mapping replacement text: \nmatched_text:%s\nreplace_term:%s", matched_text, replace_term))
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
