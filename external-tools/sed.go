package externaltools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	utils "spectre-gui/utils"
)

func Replace(row int, col int, path string, search_term string, replace_term string) error {
	utils.Log(fmt.Sprintf("replacing: %s with %s", search_term, replace_term))
	regex := fmt.Sprintf(
		`%ds/^(.{%d})%s/\1%s/`,
		row,
		col-1,
		search_term,
		// HACK: this shouldnt be necessary
		// where is the line break coming from?
		strings.Trim(replace_term, "\n"),
	)

	cmd := exec.Command(
		"sed",
		"-i",
		"-E",
		regex,
		path,
	)
	utils.Log(fmt.Sprintf("sed command: %s", cmd.String()))
	return cmd.Run()
}

func ReplaceLine(path string, row int, replacement string) error {
	regex := fmt.Sprintf(`%dc\%s`, row, replacement)
	cmd := exec.Command("sed", "-i", regex, path)
	err := cmd.Run()
	if err != nil {
		utils.Log(err.Error())
		return err
	}
	return nil
}

func GetLine(path string, row int) (string, error) {
	regex := fmt.Sprintf("%dp", row)
	cmd := exec.Command("sed", "-n", regex, path)
	bytes, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(bytes), nil
}

func GetReplacementText(
	matched_line string,
	search_term string,
	replace_term string,
	use_regex bool,
) (string, error) {
	cmd_echo := exec.Command("echo", matched_line)
	escaped := EscapeSpecialChars(search_term, use_regex)
	utils.Log(fmt.Sprintf("search term: %s escaped:  %s", search_term, escaped))
	regex := fmt.Sprintf("s/.*%s.*/%s/ip", escaped, replace_term)
	cmd_sed := exec.Command("sed", "-n", "-E", regex)
	utils.Log(fmt.Sprintf("sed command: %s", cmd_sed.String()))
	var output bytes.Buffer
	cmd_sed.Stdout = &output
	cmd_sed.Stdin, _ = cmd_echo.StdoutPipe()

	if err := cmd_sed.Start(); err != nil {
		utils.Log("[GetReplacementText] Error starting sed command:")
		utils.Log(err.Error())
		return "", err
	}

	if err := cmd_echo.Run(); err != nil {
		utils.Log("[GetReplacementText] Error running echo command:")
		utils.Log(err.Error())
		return "", err
	}

	if err := cmd_sed.Wait(); err != nil {
		utils.Log("[GetReplacementText]E rror waiting for sed command:")
		utils.Log(err.Error())
		return "", err
	}
	return strings.Trim(output.String(), "\n"), nil
}

func EscapeSpecialChars(search_term string, use_regex bool) string {
	specialChars := []string{`/`, `&`}
	if !use_regex {
		specialChars = append(specialChars, `(`, `)`, `\`, `.`, `*`, `$`, `[`, `]`, `^`)
	}
	escaped := search_term

	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, `\`+char)
	}

	return escaped
}
