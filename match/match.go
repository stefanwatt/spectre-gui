package match

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	ext "spectre-gui/external-tools"
	"spectre-gui/utils"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MatchesOfFile struct {
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
	Html            string
}

func MapSearchResult(matches []Match) []MatchesOfFile {
	grouped := make(map[string][]Match)
	for _, match := range matches {
		key := match.AbsolutePath
		grouped[key] = append(grouped[key], match)
	}
	var search_results []MatchesOfFile
	for key, value := range grouped {
		shortPath := key
		search_results = append(search_results, MatchesOfFile{Path: shortPath, Matches: value})
	}
	search_results = map_unique_paths(search_results)
	sort.Slice(search_results, func(i, j int) bool {
		return search_results[i].Path > search_results[j].Path
	})
	return search_results
}

func GetTotalMatches(grouped_matches []MatchesOfFile) int {
	total := 0
	for _, group := range grouped_matches {
		total += len(group.Matches)
	}
	return total
}

func map_unique_paths(results []MatchesOfFile) []MatchesOfFile {
	updated_results := make([]MatchesOfFile, len(results))
	copy(updated_results, results)
	updated_results = utils.MapArray(updated_results, func(result MatchesOfFile) MatchesOfFile {
		same_path_results := utils.Filter(updated_results, func(found_result MatchesOfFile) bool {
			return found_result.Path == result.Path
		})
		if len(same_path_results) == 0 {
			return result
		}
		adapted_path := utils.GetLastSubdirAndFilename(result.Path)
		return MatchesOfFile{
			Path:    adapted_path,
			Matches: result.Matches,
		}
	})
	return updated_results
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
	replacement_text, err := ext.GetReplacementText(matched_line, search_term, case_corrected_replace_term, use_regex)
	if err != nil {
		replacement_text = case_corrected_replace_term
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

func MapDirs(search_results []MatchesOfFile) []string {
	var dirs []string
	for _, result := range search_results {
		group, err := utils.Find(search_results, func(found_result MatchesOfFile) bool {
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
	titleCaser := cases.Title(language.English)
	// ALL UPPERCASE
	if matched_text == "" {
		return replace_term
	}
	if matched_text == strings.ToUpper(matched_text) {
		value := strings.ToUpper(replace_term)
		return value
	}
	// FIRST LETTER UPPER
	if len(matched_text) > 0 && unicode.IsUpper(rune(matched_text[0])) && matched_text[1:] == strings.ToLower(matched_text[1:]) {
		value := titleCaser.String(replace_term)
		return value
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
