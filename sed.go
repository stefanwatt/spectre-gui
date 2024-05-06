package main

import (
	"fmt"
	"os/exec"
)

func Sed(match RipgrepMatch, search_term string, replace_term string) {
	cmd := exec.Command(
		"sed",
		"-i",
		"-e",
		fmt.Sprintf(
			`%ds/^\(.\{%d\}\)%s/\1%s/`,
			match.Row,
			match.Col-1,
			search_term,
			replace_term,
		),
		match.AbsolutePath,
	)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
