package picker

import (
	"fmt"
	"nvim-gui/neovim"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

var helpTags []*neovim.HelpTag

func (p *Picker) FindHelp(query string) []*PickerResult {
	if len(helpTags) == 0 {
		var err error
		helpTags, err = neovim.GetHelpTags()
		if err != nil {
			log.Error("error getting help tags", "error", err)
			return nil
		}
	}

	var helpTagLines []string
	for index, helpTag := range helpTags {
		line := fmt.Sprintf("%d:%s:%s:%s:%s", index, helpTag.Tag, helpTag.Filepath, helpTag.Cmd, helpTag.Lang)
		helpTagLines = append(helpTagLines, line)
	}
	filteredLines, err := filterWithFzf(helpTagLines, query)

	if err != nil {
		panic("FindHelp error filtering with fzf")
	}

	var results []*PickerResult
	for _, line := range filteredLines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 5)
		assert(len(parts) == 5, "FindHelp malformed fzf line")
		index, _ := strconv.Atoi(parts[0])
		helpTag := helpTags[index]

		relativePath, _ := filepath.Rel(filepath.Dir(helpTag.Filepath), helpTag.Filepath)
		icon, iconColor := neovim.GetFileIcon(helpTag.Filename)

		results = append(results, &PickerResult{
			Filename:     helpTag.Filename,
			RelativePath: relativePath,
			AbsolutePath: helpTag.Filepath,
			Icon:         icon,
			IconColor:    iconColor,
			Text:         helpTag.Tag,
			Row:          neovim.ResolveTagLine(helpTag.Filepath, helpTag.Cmd),
		})
	}

	if err != nil {
		panic("error getting help files")
	}
	return results
}
