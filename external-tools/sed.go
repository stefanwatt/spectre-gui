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
		fmt.Println(err)
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

func GetReplacementText(matched_line string, search_term string, replace_term string) (string, error) {
	cmd_echo := exec.Command("echo", matched_line)
	cmd_sed := exec.Command("sed", "-n", "-E", fmt.Sprintf("s/.*%s.*/%s/ip", search_term, replace_term))
	var output bytes.Buffer
	cmd_sed.Stdout = &output
	cmd_sed.Stdin, _ = cmd_echo.StdoutPipe()

	if err := cmd_sed.Start(); err != nil {
		utils.Log("Error starting sed command:")
		utils.Log(err.Error())
		return "", err
	}

	if err := cmd_echo.Run(); err != nil {
		utils.Log("Error running echo command:")
		utils.Log(err.Error())
		return "", err
	}

	if err := cmd_sed.Wait(); err != nil {
		utils.Log("Error waiting for sed command:")
		utils.Log(err.Error())
		return "", err
	}
	return strings.Trim(output.String(), "\n"), nil
}
