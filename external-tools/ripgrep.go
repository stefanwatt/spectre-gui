package externaltools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	utils "spectre-gui/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Ripgrep(
	ctx context.Context,
	search_term string,
	replace_term string,
	dir string,
	include string,
	exclude string,
	case_sensitive bool,
	regex bool,
	match_whole_word bool,
	preserve_case bool,
) ([]string, error) {
	args := map_rg_args(
		search_term,
		dir,
		include,
		exclude,
		case_sensitive,
		regex,
		match_whole_word,
	)
	cmd := exec.CommandContext(ctx, "rg", args...)
	var output bytes.Buffer
	cmd.Stdout = &output

	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	rg_output_lines := strings.Split(output.String(), "\n")
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
		utils.Log("Error parsing line:", output)
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

func map_replacement_text_preserve_case(matched_text string, replace_term string) string {
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

func map_rg_args(
	search_term string,
	dir string,
	include string,
	exclude string,
	case_sensitive bool,
	regex bool,
	match_whole_word bool,
) []string {
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

	if case_sensitive {
		args = append(args, "--case-sensitive")
	} else {
		args = append(args, "--smart-case")
	}

	if match_whole_word {
		args = append(args, "--word-regexp")
	}
	if !regex {
		args = append(args, "--fixed-strings")
	}
	args = append(args, search_term)
	args = append(args, dir)
	return args
}
