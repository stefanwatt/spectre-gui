package main

import (
	"fmt"
	"nvim-gui/neovim"
	"nvim-gui/picker"
	ext "nvim-gui/picker/external-tools"
	"nvim-gui/utils"
	"strings"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)


func (a *App) SendKey(key string, ctrl bool, alt bool, shift bool, keymapMode string) {
	utils.Log(fmt.Sprintf("SendKey key=%s ctrl=%t alt=%t shift=%t mode=%s", key, ctrl, alt, shift, keymapMode))
	switch keymapMode {
	// case "cmdline":
	case "live-grep":
		if key == "ArrowLeft" && ctrl && !alt && !shift {
			utils.Log("live-grep-prev-page")
			results := a.liveGrepPicker.GetPrevPage()
			Runtime.EventsEmit(a.ctx, "live-grep-prev-page", results, a.liveGrepPicker.State.Pagination.PageIndex)
		}
		if key == "ArrowRight" && ctrl && !alt && !shift {
			utils.Log("live-grep-next-page")
			results := a.liveGrepPicker.GetNextPage()
			Runtime.EventsEmit(a.ctx, "live-grep-next-page", results, a.liveGrepPicker.State.Pagination.PageIndex)
		}
		if key == "Enter" && !ctrl && !alt && !shift {
			Runtime.EventsEmit(a.ctx, "live-grep-open-selected-match")
		}
	default:
		err := neovim.SendKey(key, ctrl, alt, shift)
		if err != nil {
			utils.Log(err.Error())
		}
	}
}

func (a *App) FindReferences(query string) []*picker.PickerResult {
	activeWindow := neovim.NvimScreen.GetActiveWindow()
	hasCursor := activeWindow != nil &&
		activeWindow.Cursor != nil
	hasBuffer := activeWindow!= nil && activeWindow.Buffer != nil
	var results []*picker.PickerResult
	if hasCursor &&
		activeWindow.Cursor.Col == a.referencesPicker.Col &&
		activeWindow.Cursor.Row == a.referencesPicker.Row &&
		hasBuffer &&
		activeWindow.Buffer.Filepath == a.referencesPicker.Filepath {
		_results, err := a.referencesPicker.FindReferences(a.referencesPicker.References, query)
		if err != nil {
			panic("error getting references\n"+err.Error())
		}
		results = _results
	} else {
		references := neovim.GetReferencesUnderCursor()
		_results, err := a.referencesPicker.FindReferences(references, query)
		if err != nil {
			panic("error getting references\n"+err.Error())
		}
		results = _results
	}
	if hasCursor {
		a.referencesPicker.Col = activeWindow.Cursor.Col
		a.referencesPicker.Row = activeWindow.Cursor.Row
	}
	if hasBuffer {
		a.referencesPicker.Filepath = activeWindow.Buffer.Filepath
	}
	return results
}

func (a *App) OpenFile(path string, row int, col int) {
	utils.Log(fmt.Sprintf("open neovim file path=%s row=%d col=%d", path, row, col))
	err := neovim.OpenFileAt(path, row, col)
	if err != nil {
		utils.Log(err.Error())
	}
	Runtime.EventsEmit(a.ctx, "hide-live-rep")
}

func (a *App) SubstituteJump() {
	neovim.HandleSubstituteJump()
}

func (a *App) LiveGrep(
	search_term string,
	dir string,
	exclude string,
	include string,
	case_sensitive bool,
	regex bool,
	match_whole_word bool,
) picker.SearchResult {
	return a.liveGrepPicker.LiveGrep(
		search_term,
		dir,
		exclude,
		include,
		case_sensitive,
		regex,
		match_whole_word,
		a.ctx,
	)
}

func (a *App) CreateQuickfixList(entries []*neovim.QuickfixEntry) {
	neovim.SetQuickfixList(entries)
}

func (a *App) FindFiles(query string) []*picker.PickerResult {
	if strings.TrimSpace(query) == "" {
		return []*picker.PickerResult{}
	}
	//TODO: suboptimal to call getcwd on every request
	cwd := neovim.GetCwd()
	dir := utils.GetGitRepoRoot(cwd)
	results, err := picker.FindFiles(dir, query)
	if err != nil {
		utils.Log("FindFiles error getting results:\n", err.Error())
		return []*picker.PickerResult{}
	}
	for _, result := range results {
		result.AbsolutePath = dir + "/" + result.RelativePath
	}
	return results
}

func (a *App) GetReplacementText(matchedLine string, searchTerm string, replacementText string, useRegex bool) string {
	replacementText, err := ext.GetReplacementText(matchedLine, searchTerm, replacementText, useRegex)
	if err != nil {
		return ""
	}
	return replacementText
}

func (a *App) GetLiveGrepOpts() picker.LiveGrepPickerState {
	return a.liveGrepPicker.State
}

func (a *App) GetPrevPage() picker.SearchResult {
	return a.liveGrepPicker.GetPrevPage()
}

func (a *App) GetNextPage() picker.SearchResult {
	return a.liveGrepPicker.GetPrevPage()
}
