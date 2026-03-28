package picker

import (
	"bytes"
	"nvim-gui/neovim"
	"nvim-gui/utils"
	"os/exec"
	"path/filepath"
	"strings"
)

func FindFiles(dir, query string) ([]*PickerResult, error) {
	git := exec.Command("git", "-C", dir, "ls-files")
	fzf := exec.Command("fzf", "--filter="+query)

	var out bytes.Buffer
	fzf.Stdout = &out

	gitOut, err := git.StdoutPipe()
	if err != nil {
		return []*PickerResult{}, err
	}
	fzf.Stdin = gitOut

	if err := git.Start(); err != nil {
		return []*PickerResult{}, err
	}
	if err := fzf.Start(); err != nil {
		return []*PickerResult{}, err
	}
	if err := git.Wait(); err != nil {
		return []*PickerResult{}, err
	}
	if err := fzf.Wait(); err != nil {
		return []*PickerResult{}, err
	}

	lines := strings.Split(out.String(), "\n")
	lines = lines[:len(lines)-1]
	results := utils.MapArray(lines, func(line string) *PickerResult {
		filename := filepath.Base(line)
		icon, iconColor := neovim.GetFileIcon(filename)
		return &PickerResult{
			Filename:     filename,
			RelativePath: line,
			Icon:         icon,
			IconColor:    iconColor,
		}
	})

	return results, nil
}
