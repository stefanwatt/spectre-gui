package picker

import (
	"bytes"
	"fmt"
	"nvim-gui/neovim"
	"nvim-gui/utils"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type LspReferencesPicker struct {
	Col        int
	Row        int
	Filepath   string
	References []*neovim.LspReferenceItem
}

func NewReferencesPicker() *LspReferencesPicker {
	return &LspReferencesPicker{
		Col:        1,
		Row:        1,
		Filepath:   "",
		References: []*neovim.LspReferenceItem{},
	}
}

func (refPicker *LspReferencesPicker) FindReferences(references []*neovim.LspReferenceItem, query string) ([]*PickerResult, error) {
	refPicker.References = references
	if len(references) == 0 {
		return []*PickerResult{},nil
	}
	var referenceLines []string
	for _, ref := range references {
		line := fmt.Sprintf("%s:%d:%d:%s", ref.AbsolutePath, ref.StartRow, ref.StartCol, ref.Text)
		referenceLines = append(referenceLines, line)
	}
	input := strings.Join(referenceLines, "\n")
	utils.Log("FindReferences references:", references)
	utils.Log("FindReferences input:", input)
	fzf := exec.Command("fzf", "--filter="+query, "--delimiter=:", "--with-nth=1,2,3,4")
	fzf.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	fzf.Stdout = &out
	if err := fzf.Run(); err != nil {
		return []*PickerResult{}, err
	}
	selectedLines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var results []*PickerResult
	for _, line := range selectedLines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			continue
		}
		absolutePath := parts[0]
		row, _ := strconv.Atoi(parts[1])
		col, _ := strconv.Atoi(parts[2])
		text := parts[3]
		filename := filepath.Base(absolutePath)
		relativePath, _ := filepath.Rel(filepath.Dir(absolutePath), absolutePath)
		icon, iconColor := neovim.GetFileIcon(filename)
		results = append(results, &PickerResult{
			Filename:     filename,
			RelativePath: relativePath,
			AbsolutePath: absolutePath,
			Icon:         icon,
			IconColor:    iconColor,
			Text:         text,
			Row:          row,
			Col:          col,
		})
	}
	return results, nil
}
