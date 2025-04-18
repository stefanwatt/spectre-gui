package picker

import (
	"fmt"
	"nvim-gui/neovim"
	"nvim-gui/utils"
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
		return []*PickerResult{}, nil
	}

	var referenceLines []string
	for index, ref := range references {
		line := fmt.Sprintf("%d:%s:%d:%d:%s", index, ref.AbsolutePath, ref.StartRow, ref.StartCol, ref.Text)
		referenceLines = append(referenceLines, line)
	}

	utils.Log("FindReferences references:", references)
	utils.Log("FindReferences input:", strings.Join(referenceLines, "\n"))

	selectedLines, err := filterWithFzf(referenceLines, query)
	if err != nil {
		return []*PickerResult{}, err
	}

	var results []*PickerResult
	for _, line := range selectedLines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 5)
		assert(len(parts) == 5, "FindReferences malformed fzf line")
		index, _ := strconv.Atoi(parts[0])
		ref := references[index]

		filename := filepath.Base(ref.AbsolutePath)
		relativePath, _ := filepath.Rel(filepath.Dir(ref.AbsolutePath), ref.AbsolutePath)
		icon, iconColor := neovim.GetFileIcon(filename)

		results = append(results, &PickerResult{
			Filename:     filename,
			RelativePath: relativePath,
			AbsolutePath: ref.AbsolutePath,
			Icon:         icon,
			IconColor:    iconColor,
			Text:         ref.Text,
			Row:          ref.StartRow,
			Col:          ref.StartCol,
		})
	}

	return results, nil
}
