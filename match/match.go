package match

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	ext "spectre-gui/external-tools"
	"spectre-gui/utils"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SearchResult struct {
	Path    string
	Matches []Match
}

type Match struct {
	Id              string
	FileName        string
	AbsolutePath    string
	MatchedLine     string
	TextBeforeMatch string
	TextAfterMatch  string
	MatchedText     string
	ReplacementText string
	Row             int
	Col             int
}

func MapSearchResult(matches []Match) []SearchResult {
	grouped := make(map[string][]Match)
	for _, match := range matches {
		key := match.FileName
		grouped[key] = append(grouped[key], match)
	}
	var search_results []SearchResult
	for key, value := range grouped {
		search_results = append(search_results, SearchResult{Path: key, Matches: value})
	}
	return search_results
}

func MapMatch(
	output string,
	path string,
	matched_text string,
	row int,
	col int,
	search_term string,
	replace_term string,
	preserve_case bool,
	use_regex bool,
) Match {
	matched_line, err := ext.GetLine(path, row)
	if err != nil {
		panic(err)
	}
	uuid, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	case_corrected_replace_term := replace_term
	if preserve_case {
		case_corrected_replace_term = map_replacement_text_preserve_case(matched_text, replace_term)
	}
	replacement_text, err := ext.GetReplacementText(matched_line, search_term, case_corrected_replace_term)
	if err != nil {
		panic(err)
	}
	before, after := map_before_and_after(matched_line, matched_text)
	match := Match{
		Id:              uuid.String(),
		FileName:        filepath.Base(path),
		AbsolutePath:    path,
		MatchedLine:     matched_line,
		TextBeforeMatch: before,
		TextAfterMatch:  after,
		MatchedText:     matched_text,
		ReplacementText: replacement_text,
		Row:             row,
		Col:             col,
	}
	return match
}

func MapDirs(search_results []SearchResult) []string {
	var dirs []string
	for _, result := range search_results {
		group, err := utils.Find(search_results, func(found_result SearchResult) bool {
			return found_result.Path == result.Path
		})
		if err != nil {
			panic("dir not found")
		}
		current_dir := filepath.Dir(group.Matches[0].AbsolutePath)
		_, err = utils.Find(dirs, func(dir string) bool {
			return dir == current_dir
		})
		if err != nil {
			dirs = append(dirs, current_dir)
		}
	}
	return dirs
}

func map_replacement_text_preserve_case(matched_text string, replace_term string) string {
	utils.Log(fmt.Sprintf("mapping replacement text: \nmatched_text:%s\nreplace_term:%s", matched_text, replace_term))
	titleCaser := cases.Title(language.English)
	// ALL UPPERCASE
	if matched_text == "" {
		return strings.Trim(replace_term, "\n")
	}
	if matched_text == strings.ToUpper(matched_text) {
		value := strings.ToUpper(replace_term)
		return strings.Trim(value, "\n")
	}
	// FIRST LETTER UPPER
	if len(matched_text) > 0 && unicode.IsUpper(rune(matched_text[0])) && matched_text[1:] == strings.ToLower(matched_text[1:]) {
		value := titleCaser.String(replace_term)
		return strings.Trim(value, "\n")
	}
	// DEFAULT
	return strings.Trim(replace_term, "\n")
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
