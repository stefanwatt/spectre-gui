package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"strings"
)

type FileExplorerConfirmPrompt struct {
	WinId   int      `json:"winId"`
	Message string   `json:"message"`
	Choices []string `json:"choices"`
}

func detectFileExplorerConfirmPrompt(window *Window) (*FileExplorerConfirmPrompt, bool) {
	if window == nil || window.Buffer == nil || window.Grid == nil {
		return nil, false
	}
	if (*window.Buffer).Filetype != "" {
		return nil, false
	}
	raw := strings.TrimSpace(window.Grid.toString())
	if raw == "" {
		return nil, false
	}
	if !strings.Contains(raw, "without synchronization") {
		return nil, false
	}
	if !strings.Contains(raw, "Confirm") {
		return nil, false
	}

	// Extract text from internal confirm UI by dropping decorative borders and option markers.
	lines := strings.Split(raw, "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "│") || strings.Contains(trimmed, "─") || strings.Contains(trimmed, "┌") || strings.Contains(trimmed, "└") {
			trimmed = strings.Map(func(r rune) rune {
				switch r {
				case '│', '─', '┌', '┐', '└', '┘':
					return -1
				default:
					return r
				}
			}, trimmed)
			trimmed = strings.TrimSpace(trimmed)
		}
		if trimmed == "" {
			continue
		}
		trimmed = strings.ReplaceAll(trimmed, "&", "")
		if trimmed == "Yes" || trimmed == "No" || strings.HasPrefix(trimmed, "Yes") || strings.HasPrefix(trimmed, "No") {
			continue
		}
		cleanLines = append(cleanLines, trimmed)
	}

	message := strings.Join(cleanLines, "\n")
	if strings.TrimSpace(message) == "" {
		message = "Confirm close without synchronization?"
	}

	return &FileExplorerConfirmPrompt{
		WinId:   window.ID,
		Message: message,
		Choices: []string{"Yes", "No"},
	}, true
}

func (s *Screen) emitFileExplorerConfirmPrompt(window *Window) bool {
	prompt, ok := detectFileExplorerConfirmPrompt(window)
	if !ok {
		return false
	}
	s.emitEvent("file-explorer-confirm-prompt-show", prompt)
	return true
}

func (s *Screen) HandleFileExplorerConfirmChoice(winId, choice int) {
	if NvimInstance == nil || winId < 1 {
		return
	}
	var key string
	switch choice {
	case 1:
		key = "<CR>"
	default:
		key = "<Esc>"
	}
	lua := `
local win_id, feed = ...
if not vim.api.nvim_win_is_valid(win_id) then
  return false
end
vim.api.nvim_set_current_win(win_id)
vim.api.nvim_feedkeys(vim.api.nvim_replace_termcodes(feed, true, false, true), "n", false)
return true
`
	var ok bool
	if err := NvimInstance.ExecLua(lua, &ok, winId, key); err != nil {
		utils.Log(fmt.Sprintf("[minifiles] failed sending confirm response to win=%d: %v", winId, err))
		return
	}
	if ok {
		s.emitEvent("file-explorer-confirm-prompt-hide", struct{}{})
	}
}
