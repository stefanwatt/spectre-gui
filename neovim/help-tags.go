package neovim

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

type HelpTag struct {
	Tag      string `msgpack:"tag"`
	Filename string `msgpack:"filename"`
	Filepath string `msgpack:"filepath"`
	Cmd      string `msgpack:"cmd"`
	Lang     string `msgpack:"lang"`
}

func GetHelpTags() ([]*HelpTag, error) {
	var runtimePaths []string
	err := NvimClient.ExecLua("return vim.api.nvim_list_runtime_paths()", &runtimePaths)
	if err != nil {
		return nil, fmt.Errorf("error getting runtime paths: %w", err)
	}

	var helpTags []*HelpTag
	for _, rtp := range runtimePaths {
		docDir := filepath.Join(rtp, "doc")
		entries, err := os.ReadDir(docDir)
		if err != nil {
			continue // doc dir doesn't exist or isn't readable
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, "tags") {
				continue
			}
			lang := ""
			if strings.HasPrefix(name, "tags-") {
				lang = strings.TrimPrefix(name, "tags-")
			} else if name != "tags" {
				continue
			}

			tagsFile := filepath.Join(docDir, name)
			tags, err := parseTagsFile(tagsFile, docDir, lang)
			if err != nil {
				log.Warn("error parsing tags file", "path", tagsFile, "error", err)
				continue
			}
			helpTags = append(helpTags, tags...)
		}
	}

	log.Debug("GetHelpTags", "count", len(helpTags))
	return helpTags, nil
}

// ResolveTagLine searches for the tag pattern in the help file and returns
// the 1-based line number where it occurs. The cmd field from a tags file is
// typically "/*tag-name*", meaning search for literal "*tag-name*" in the file.
// Returns 1 if the tag cannot be found.
func ResolveTagLine(filepath string, cmd string) int {
	// Extract the search pattern from the cmd field.
	// Format is /*pattern* — strip the leading / and use the rest as a literal search.
	pattern := strings.TrimPrefix(cmd, "/")
	// The pattern is wrapped in \* anchors for :tag, but in the file it appears literally.
	// e.g. cmd="/*lua-guide*" → search for "*lua-guide*" in the file.

	file, err := os.Open(filepath)
	if err != nil {
		return 1
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if strings.Contains(scanner.Text(), pattern) {
			return lineNum
		}
	}
	return 1
}

func parseTagsFile(tagsFile, docDir, lang string) ([]*HelpTag, error) {
	file, err := os.Open(tagsFile)
	if err != nil {
		return nil, fmt.Errorf("error opening tags file: %w", err)
	}
	defer file.Close()

	var tags []*HelpTag
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "!") {
			continue // skip empty lines and tag file metadata
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}

		tag := parts[0]
		filename := parts[1]
		cmd := parts[2]

		tags = append(tags, &HelpTag{
			Tag:      tag,
			Filename: filename,
			Filepath: filepath.Join(docDir, filename),
			Cmd:      cmd,
			Lang:     lang,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading tags file: %w", err)
	}

	return tags, nil
}
