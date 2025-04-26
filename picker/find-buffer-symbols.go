package picker

import (
	"fmt"
	"nvim-gui/neovim"
	"nvim-gui/utils"
	"strconv"
	"strings"
)

var symbolIcons = map[string]string{
	"File":          "󰈙",
	"Module":        "󰕳",
	"Namespace":     "󰦮",
	"Package":       "",
	"Class":         "󰆧",
	"Method":        "󰊕",
	"Property":      "",
	"Field":         "",
	"Constructor":   "󰅩",
	"Enum":          "",
	"Interface":     "",
	"Function":      "󰊕",
	"Variable":      "󰀫",
	"Constant":      "󰏿",
	"String":        "",
	"Number":        "󰎠",
	"Boolean":       "󰨙",
	"Array":         "󱡠",
	"Object":        "󰅩",
	"Key":           "󰌋",
	"Null":          "󰟢",
	"EnumMember":    "",
	"Struct":        "󰆼",
	"Event":         "",
	"Operator":      "󰆕",
	"TypeParameter": "󰗴",
}

type LspSymbolsPicker struct {
	Filepath string
	Symbols  []*neovim.LspSymbolItem
}

func NewSymbolsPicker() *LspSymbolsPicker {
	return &LspSymbolsPicker{
		Filepath: "",
		Symbols:  []*neovim.LspSymbolItem{},
	}
}

func (sp *LspSymbolsPicker) FindBufferSymbols(query string) []*PickerResult {
	activeWindow := neovim.NvimScreen.GetActiveWindow()
	hasBuffer := activeWindow != nil && activeWindow.Buffer != nil
	var results []*PickerResult
	if hasBuffer &&
		activeWindow.Buffer.Filepath == sp.Filepath {
		_results, err := sp.filterMapSymbols(sp.Symbols, query)
		if err != nil {
			panic("error getting references\n" + err.Error())
		}
		results = _results
	} else {
		symbols := neovim.GetDocumentSymbols()
		_results, err := sp.filterMapSymbols(symbols, query)
		if err != nil {
			panic("error getting references\n" + err.Error())
		}
		results = _results
	}
	if hasBuffer {
		sp.Filepath = activeWindow.Buffer.Filepath
	}
	return results
}

func (symbolPicker *LspSymbolsPicker) filterMapSymbols(symbols []*neovim.LspSymbolItem, query string) ([]*PickerResult, error) {
	symbolPicker.Symbols = symbols
	if len(symbols) == 0 {
		return []*PickerResult{}, nil
	}

	var symbolLines []string
	for index, sym := range symbols {
		line := fmt.Sprintf("%d:%s:%s", index, sym.Text, sym.Kind)
		symbolLines = append(symbolLines, line)
	}

	utils.Log("FindSymbols symbols:", symbols)
	utils.Log("FindSymbols input:", strings.Join(symbolLines, "\n"))

	selectedLines, err := filterWithFzf(symbolLines, query)
	if err != nil {
		return []*PickerResult{}, err
	}

	var results []*PickerResult
	for _, line := range selectedLines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			panic("FindSymbols malformed fzf entry")
		}

		index, _ := strconv.Atoi(parts[0])
		symbol := symbols[index]
		kindIcon, exists := symbolIcons[symbol.Kind]
		icon := ""
		if exists {
			icon = kindIcon
		}

		results = append(results, &PickerResult{
			AbsolutePath: symbol.AbsolutePath,
			Icon:         icon,
			IconColor:    "black",
			Text:         removeBracketedText(symbol.Text),
			Row:          symbol.StartRow,
			Col:          symbol.StartCol,
			// You might want to add Kind field to PickerResult if needed
		})
	}

	return results, nil
}

func removeBracketedText(s string) string {
	startIndex := strings.Index(s, "[")
	if startIndex == -1 {
		return s
	}

	endIndex := strings.Index(s, "]")
	if endIndex == -1 || endIndex < startIndex {
		return s
	}
	return strings.TrimSpace(strings.TrimSpace(s[:startIndex]) + " " + strings.TrimSpace(s[endIndex+1:]))
}
