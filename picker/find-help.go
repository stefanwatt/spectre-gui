package picker

import (
	"fmt"
	"nvim-gui/neovim"
	"path/filepath"
	"strconv"
	"strings"
)

var helpTags []*neovim.HelpTag

func FindHelp(query string) ([]*PickerResult, error) {
	if len(helpTags) == 0 {
		helpTags = neovim.GetHelpTags()
	}

	var helpTagLines []string
	for index, helpTag := range helpTags {
		line := fmt.Sprintf("%d:%s:%s:%s:%s", index, helpTag.Tag, helpTag.Filepath, helpTag.Cmd, helpTag.Lang)
		helpTagLines = append(helpTagLines, line)
	}
	filteredLines, err := filterWithFzf(helpTagLines, query)

	if err != nil {
		return []*PickerResult{}, err
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
		})
	}

	return results, nil
}
