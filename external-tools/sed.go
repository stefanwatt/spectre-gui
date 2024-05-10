package externaltools

import (
	"bytes"
	"fmt"
	"os/exec"

	utils "spectre-gui/utils"
)

func Sed(row int, col int, path string, search_term string, replace_term string, preserve_case bool) {
	cmd := exec.Command(
		"sed",
		"-i",
		"-e",
		fmt.Sprintf(
			`%ds/^\(.\{%d\}\)%s/\1%s/`,
			row,
			col-1,
			search_term,
			replace_term,
		),
		path,
	)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func GetLine(path string, row int) string {
	cmd := exec.Command("sed", "-n", fmt.Sprintf("%dp", row), path)
	bytes, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(bytes)
}

func GetReplacementText(matched_line string, search_term string, replace_term string) string {
	utils.Log("Gettting replacement text")
	cmd_echo := exec.Command("echo", matched_line)
	cmd_sed := exec.Command("sed", "-n", "-E", fmt.Sprintf("s/.*%s.*/%s/ip", search_term, replace_term))
	var output bytes.Buffer
	cmd_sed.Stdout = &output
	cmd_sed.Stdin, _ = cmd_echo.StdoutPipe()

	if err := cmd_sed.Start(); err != nil {
		utils.Log("Error starting sed command:")
		utils.Log(err.Error())
		return ""
	}

	if err := cmd_echo.Run(); err != nil {
		utils.Log("Error running echo command:")
		utils.Log(err.Error())
		return ""
	}

	if err := cmd_sed.Wait(); err != nil {
		utils.Log("Error waiting for sed command:")
		utils.Log(err.Error())
		return ""
	}
	return output.String()
}
