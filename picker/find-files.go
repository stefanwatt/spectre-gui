package picker

import (
	"bytes"
	"nvim-gui/neovim"
	"nvim-gui/utils"
	"os/exec"
	"path/filepath"
	"strings"
)

type FindFilesResult struct {
	Filename     string `json:"filename"`
	RelativePath string `json:"relativePath"`
	AbsolutePath string `json:"absolutePath"`
	Icon         string `json:"icon"`
	IconColor    string `json:"iconColor"`
}

func FindFiles(dir, query string) ([]*FindFilesResult, error) {
	git := exec.Command("git", "-C", dir, "ls-files")
	fzf := exec.Command("fzf", "--filter="+query)

	var out bytes.Buffer
	fzf.Stdout = &out

	gitOut, err := git.StdoutPipe()
	if err != nil {
		return []*FindFilesResult{}, err
	}
	fzf.Stdin = gitOut

	if err := git.Start(); err != nil {
		return []*FindFilesResult{}, err
	}
	if err := fzf.Start(); err != nil {
		return []*FindFilesResult{}, err
	}
	if err := git.Wait(); err != nil {
		return []*FindFilesResult{}, err
	}
	if err := fzf.Wait(); err != nil {
		return []*FindFilesResult{}, err
	}

	lines := strings.Split(out.String(), "\n")
	lines = lines[:len(lines)-1]
	results := utils.MapArray(lines, func(line string) *FindFilesResult {
		filename := filepath.Base(line)
		icon, iconColor := neovim.GetFileIcon(filename)
		return &FindFilesResult{
			Filename:     filename,
			RelativePath: line,
			Icon:         icon,
			IconColor:    iconColor,
		}
	})

	return results, nil
}
