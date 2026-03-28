package picker

import (
	"context"
	"fmt"
	"nvim-gui/neovim"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/wailsapp/wails/v3/pkg/application"
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

func (rp *LspReferencesPicker) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return nil
}

func (rp *LspReferencesPicker) FindReferences(query string) []*PickerResult {
	activeWindow := neovim.NvimScreen.GetActiveWindow()
	hasCursor := activeWindow != nil &&
		activeWindow.Cursor != nil
	hasBuffer := activeWindow != nil && activeWindow.Buffer != nil
	var results []*PickerResult
	if hasCursor &&
		activeWindow.Cursor.Col == rp.Col &&
		activeWindow.Cursor.Row == rp.Row &&
		hasBuffer &&
		activeWindow.Buffer.Filepath == rp.Filepath {
		_results, err := rp.filterMapReferences(rp.References, query)
		if err != nil {
			panic("error getting references\n" + err.Error())
		}
		results = _results
	} else {
		references := neovim.GetReferencesUnderCursor()
		_results, err := rp.filterMapReferences(references, query)
		if err != nil {
			panic("error getting references\n" + err.Error())
		}
		results = _results
	}
	if hasCursor {
		rp.Col = activeWindow.Cursor.Col
		rp.Row = activeWindow.Cursor.Row
	}
	if hasBuffer {
		rp.Filepath = activeWindow.Buffer.Filepath
	}
	return results
}

func (refPicker *LspReferencesPicker) filterMapReferences(references []*neovim.LspReferenceItem, query string) ([]*PickerResult, error) {
	refPicker.References = references
	if len(references) == 0 {
		return []*PickerResult{}, nil
	}

	var referenceLines []string
	for index, ref := range references {
		line := fmt.Sprintf("%d:%s:%d:%d:%s", index, ref.AbsolutePath, ref.StartRow, ref.StartCol, ref.Text)
		referenceLines = append(referenceLines, line)
	}

	log.Debug("FindReferences references:", references)
	log.Debug("FindReferences input:", strings.Join(referenceLines, "\n"))

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
