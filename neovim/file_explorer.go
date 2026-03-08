package neovim

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"nvim-gui/utils"

	"github.com/neovim/go-client/nvim"
)

type FileExplorer struct {
	screen          *Screen
	Active          bool
	CurrentDir      string
	centerBuf       nvim.Buffer
	centerWin       nvim.Window
	centerGridId    int
	previewBuf      nvim.Buffer
	previewWin      nvim.Window
	previewGridId   int
	originalEntries []FileEntry
	namespace       int
	mu              sync.Mutex
}

type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

func NewFileExplorer(screen *Screen) *FileExplorer {
	return &FileExplorer{
		screen: screen,
	}
}

func (fe *FileExplorer) Open(dir string) {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if fe.Active {
		fe.closeLocked()
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error resolving path: %v", err))
		return
	}
	fe.CurrentDir = absDir

	entries, err := fe.readDir(absDir)
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error reading dir: %v", err))
		return
	}
	fe.originalEntries = entries

	// Create namespace for extmarks
	fe.namespace, err = NvimInstance.CreateNamespace("file_explorer")
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error creating namespace: %v", err))
		return
	}

	// Create scratch buffer for center column
	fe.centerBuf, err = NvimInstance.CreateBuffer(false, true)
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error creating center buffer: %v", err))
		return
	}

	// Set buffer lines
	lines := fe.entryLines(entries)
	err = NvimInstance.SetBufferLines(fe.centerBuf, 0, -1, false, linesToBytes(lines))
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error setting buffer lines: %v", err))
		return
	}

	// Set buffer options
	NvimInstance.SetBufferOption(fe.centerBuf, "buftype", "nofile")
	NvimInstance.SetBufferOption(fe.centerBuf, "bufhidden", "wipe")
	NvimInstance.SetBufferOption(fe.centerBuf, "filetype", "file-explorer")

	// Place extmarks on each line to track identity
	for i := range entries {
		NvimInstance.SetBufferExtmark(fe.centerBuf, fe.namespace, i, 0, map[string]interface{}{})
	}

	// Open center floating window (~50% of editor, centered)
	var editorWidth, editorHeight int
	NvimInstance.ExecLua("return vim.o.columns", &editorWidth)
	NvimInstance.ExecLua("return vim.o.lines", &editorHeight)

	centerWidth := editorWidth / 2
	centerHeight := editorHeight - 4
	centerCol := editorWidth / 4
	previewWidth := editorWidth / 4

	fe.centerWin, err = NvimInstance.OpenWindow(fe.centerBuf, true, &nvim.WindowConfig{
		Relative:  "editor",
		Width:     centerWidth,
		Height:    centerHeight,
		Row:       1,
		Col:       float64(centerCol),
		Style:     "minimal",
		ZIndex:    200,
		Focusable: true,
	})
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error opening center window: %v", err))
		return
	}

	// Create preview buffer
	fe.previewBuf, err = NvimInstance.CreateBuffer(false, true)
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error creating preview buffer: %v", err))
		return
	}
	NvimInstance.SetBufferOption(fe.previewBuf, "buftype", "nofile")
	NvimInstance.SetBufferOption(fe.previewBuf, "bufhidden", "wipe")
	NvimInstance.SetBufferOption(fe.previewBuf, "filetype", "file-explorer-preview")

	// Open preview floating window (right ~25%)
	fe.previewWin, err = NvimInstance.OpenWindow(fe.previewBuf, false, &nvim.WindowConfig{
		Relative:  "editor",
		Width:     previewWidth,
		Height:    centerHeight,
		Row:       1,
		Col:       float64(centerCol + centerWidth + 1),
		Style:     "minimal",
		ZIndex:    200,
		Focusable: false,
	})
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Open: error opening preview window: %v", err))
		return
	}

	fe.Active = true

	// Set buffer-local keymaps
	fe.setupKeymaps()

	// Set CursorMoved autocmd
	fe.setupAutocmds()

	// Emit open event with parent entries and window IDs
	parentEntries := fe.getParentEntries()
	if App != nil {
		App.Event.Emit("file-explorer-open", map[string]interface{}{
			"currentDir":      fe.CurrentDir,
			"parentEntries":   parentEntries,
			"centerWindowId":  int(fe.centerWin),
			"previewWindowId": int(fe.previewWin),
		})
	}

	// Trigger initial preview for first entry
	if len(entries) > 0 {
		fe.updatePreviewLocked(0)
	}
}

func (fe *FileExplorer) Close() {
	fe.mu.Lock()
	defer fe.mu.Unlock()
	fe.closeLocked()
}

func (fe *FileExplorer) closeLocked() {
	if !fe.Active {
		return
	}
	fe.Active = false

	// Close windows (ignore errors — may already be closed)
	NvimInstance.CloseWindow(fe.previewWin, true)
	NvimInstance.CloseWindow(fe.centerWin, true)

	if App != nil {
		App.Event.Emit("file-explorer-close", struct{}{})
	}
}

func (fe *FileExplorer) UpdatePreview(row int) {
	fe.mu.Lock()
	defer fe.mu.Unlock()
	fe.updatePreviewLocked(row)
}

func (fe *FileExplorer) updatePreviewLocked(row int) {
	if !fe.Active {
		return
	}

	// Get entry at row via extmarks
	entry := fe.getEntryAtRow(row)
	if entry == nil {
		return
	}

	if entry.IsDir {
		// Directory preview — read dir and emit as list
		subEntries, err := fe.readDir(entry.Path)
		if err != nil {
			utils.Log(fmt.Sprintf("FileExplorer.UpdatePreview: error reading dir: %v", err))
			return
		}
		// Set preview buffer to dir listing
		lines := fe.entryLines(subEntries)
		NvimInstance.SetBufferLines(fe.previewBuf, 0, -1, false, linesToBytes(lines))

		if App != nil {
			App.Event.Emit("file-explorer-preview", map[string]interface{}{
				"isFile":  false,
				"entries": subEntries,
			})
		}
	} else {
		// File preview — read file and set preview buffer with syntax
		content, err := readFilePreview(entry.Path, 200)
		if err != nil {
			NvimInstance.SetBufferLines(fe.previewBuf, 0, -1, false, linesToBytes([]string{"[cannot read file]"}))
			return
		}
		NvimInstance.SetBufferLines(fe.previewBuf, 0, -1, false, linesToBytes(content))

		// Set filetype based on extension for syntax highlighting
		ext := filepath.Ext(entry.Name)
		ft := extToFiletype(ext)
		if ft != "" {
			NvimInstance.SetBufferOption(fe.previewBuf, "filetype", ft)
		}

		if App != nil {
			App.Event.Emit("file-explorer-preview", map[string]interface{}{
				"isFile":  true,
				"entries": []FileEntry{},
			})
		}
	}
}

func (fe *FileExplorer) NavigateInto() {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if !fe.Active {
		return
	}

	// Get current cursor row
	row := fe.getCursorRow()
	entry := fe.getEntryAtRow(row)
	if entry == nil {
		return
	}

	if entry.IsDir {
		// Navigate into directory
		fe.CurrentDir = entry.Path
		fe.refreshLocked()
	} else {
		// Open file: close explorer and open in neovim
		fe.closeLocked()
		NvimInstance.Command(fmt.Sprintf("edit %s", entry.Path))
	}
}

func (fe *FileExplorer) NavigateUp() {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if !fe.Active {
		return
	}

	parent := filepath.Dir(fe.CurrentDir)
	if parent == fe.CurrentDir {
		return // already at root
	}
	fe.CurrentDir = parent
	fe.refreshLocked()
}

func (fe *FileExplorer) Apply() {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if !fe.Active {
		return
	}

	// Read current buffer lines
	currentLines, err := NvimInstance.BufferLines(fe.centerBuf, 0, -1, false)
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Apply: error reading buffer lines: %v", err))
		return
	}

	// Get all extmarks
	extmarks, err := NvimInstance.BufferExtmarks(fe.centerBuf, fe.namespace, 0, -1, map[string]interface{}{})
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.Apply: error getting extmarks: %v", err))
		return
	}

	// Build map: extmark ID → current row
	extmarkRows := make(map[int]int) // extmark ID → row
	for _, em := range extmarks {
		extmarkRows[em.ID] = em.Row
	}

	// Track which lines are accounted for by extmarks
	accountedLines := make(map[int]bool)

	// Process renames and track deletions
	for i, original := range fe.originalEntries {
		emRow, exists := extmarkRows[i+1] // extmark IDs are 1-indexed
		if !exists {
			// Extmark gone → line was deleted
			utils.Log(fmt.Sprintf("FileExplorer.Apply: deleting %s", original.Path))
			os.RemoveAll(original.Path)
			continue
		}

		accountedLines[emRow] = true
		if emRow >= len(currentLines) {
			continue
		}

		newName := strings.TrimSpace(string(currentLines[emRow]))
		newName = strings.TrimSuffix(newName, "/")
		oldName := strings.TrimSuffix(original.Name, "/")

		if newName != oldName && newName != "" {
			// Rename
			newPath := filepath.Join(fe.CurrentDir, newName)
			utils.Log(fmt.Sprintf("FileExplorer.Apply: renaming %s → %s", original.Path, newPath))
			err := os.Rename(original.Path, newPath)
			if err != nil {
				utils.Log(fmt.Sprintf("FileExplorer.Apply: rename error: %v", err))
			}
		}
	}

	// Lines without extmarks → new files/dirs
	for i, line := range currentLines {
		if accountedLines[i] {
			continue
		}
		name := strings.TrimSpace(string(line))
		if name == "" {
			continue
		}
		newPath := filepath.Join(fe.CurrentDir, name)
		if strings.HasSuffix(name, "/") {
			utils.Log(fmt.Sprintf("FileExplorer.Apply: creating dir %s", newPath))
			os.MkdirAll(newPath, 0755)
		} else {
			utils.Log(fmt.Sprintf("FileExplorer.Apply: creating file %s", newPath))
			f, err := os.Create(newPath)
			if err != nil {
				utils.Log(fmt.Sprintf("FileExplorer.Apply: create error: %v", err))
			} else {
				f.Close()
			}
		}
	}

	// Refresh
	fe.refreshLocked()
}

func (fe *FileExplorer) refreshLocked() {
	entries, err := fe.readDir(fe.CurrentDir)
	if err != nil {
		utils.Log(fmt.Sprintf("FileExplorer.refresh: error reading dir: %v", err))
		return
	}
	fe.originalEntries = entries

	// Clear and set buffer lines
	lines := fe.entryLines(entries)
	NvimInstance.SetBufferLines(fe.centerBuf, 0, -1, false, linesToBytes(lines))

	// Clear existing extmarks and set new ones
	NvimInstance.ClearBufferNamespace(fe.centerBuf, fe.namespace, 0, -1)
	for i := range entries {
		NvimInstance.SetBufferExtmark(fe.centerBuf, fe.namespace, i, 0, map[string]interface{}{})
	}

	// Emit updated state
	parentEntries := fe.getParentEntries()
	if App != nil {
		App.Event.Emit("file-explorer-open", map[string]interface{}{
			"currentDir":      fe.CurrentDir,
			"parentEntries":   parentEntries,
			"centerWindowId":  int(fe.centerWin),
			"previewWindowId": int(fe.previewWin),
		})
	}

	// Update preview for first entry
	if len(entries) > 0 {
		fe.updatePreviewLocked(0)
	}
}

func (fe *FileExplorer) setupKeymaps() {
	ch := NvimInstance.ChannelID()
	luaCode := fmt.Sprintf(`
		local buf = vim.api.nvim_get_current_buf()
		local opts = { buffer = buf, noremap = true, silent = true }
		vim.keymap.set('n', 'h', function() vim.rpcnotify(%d, 'fe-navigate-up') end, opts)
		vim.keymap.set('n', 'l', function() vim.rpcnotify(%d, 'fe-navigate-into') end, opts)
		vim.keymap.set('n', '<CR>', function() vim.rpcnotify(%d, 'fe-navigate-into') end, opts)
		vim.keymap.set('n', 'q', function() vim.rpcnotify(%d, 'fe-close') end, opts)
		vim.keymap.set('n', '<Esc>', function() vim.rpcnotify(%d, 'fe-close') end, opts)
		vim.keymap.set('n', '=', function() vim.rpcnotify(%d, 'fe-apply') end, opts)
	`, ch, ch, ch, ch, ch, ch)
	NvimInstance.ExecLua(luaCode, nil)
}

func (fe *FileExplorer) setupAutocmds() {
	ch := NvimInstance.ChannelID()
	luaCode := fmt.Sprintf(`
		local buf = vim.api.nvim_get_current_buf()
		vim.api.nvim_create_autocmd('CursorMoved', {
			buffer = buf,
			callback = function()
				local row = vim.api.nvim_win_get_cursor(0)[1] - 1
				vim.rpcnotify(%d, 'fe-cursor-moved', row)
			end,
		})
	`, ch)
	NvimInstance.ExecLua(luaCode, nil)
}

func (fe *FileExplorer) getCursorRow() int {
	var pos [2]int
	err := NvimInstance.ExecLua("return vim.api.nvim_win_get_cursor(0)", &pos)
	if err != nil {
		return 0
	}
	return pos[0] - 1 // convert 1-indexed to 0-indexed
}

func (fe *FileExplorer) getEntryAtRow(row int) *FileEntry {
	if row < 0 || row >= len(fe.originalEntries) {
		return nil
	}
	return &fe.originalEntries[row]
}

func (fe *FileExplorer) getParentEntries() []FileEntry {
	parent := filepath.Dir(fe.CurrentDir)
	entries, err := fe.readDir(parent)
	if err != nil {
		return nil
	}
	return entries
}

func (fe *FileExplorer) readDir(dir string) ([]FileEntry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entries []FileEntry
	for _, de := range dirEntries {
		// Skip hidden files
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}
		entries = append(entries, FileEntry{
			Name:  de.Name(),
			Path:  filepath.Join(dir, de.Name()),
			IsDir: de.IsDir(),
		})
	}

	// Sort: dirs first, then files, alphabetical within each group
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	return entries, nil
}

func (fe *FileExplorer) entryLines(entries []FileEntry) []string {
	lines := make([]string, len(entries))
	for i, e := range entries {
		if e.IsDir {
			lines[i] = e.Name + "/"
		} else {
			lines[i] = e.Name
		}
	}
	return lines
}

// Helper functions

func linesToBytes(lines []string) [][]byte {
	result := make([][]byte, len(lines))
	for i, l := range lines {
		result[i] = []byte(l)
	}
	return result
}

func readFilePreview(path string, maxLines int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Check if file appears to be binary
	if isBinaryContent(data) {
		return []string{"[binary file]"}, nil
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return lines, nil
}

func isBinaryContent(data []byte) bool {
	// Check first 512 bytes for null bytes
	check := data
	if len(check) > 512 {
		check = check[:512]
	}
	for _, b := range check {
		if b == 0 {
			return true
		}
	}
	return false
}

func extToFiletype(ext string) string {
	ftMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "javascriptreact",
		".tsx":  "typescriptreact",
		".py":   "python",
		".rs":   "rust",
		".lua":  "lua",
		".md":   "markdown",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".toml": "toml",
		".html": "html",
		".css":  "css",
		".svelte": "svelte",
		".sh":   "sh",
		".bash": "bash",
		".nix":  "nix",
		".vim":  "vim",
		".c":    "c",
		".h":    "c",
		".cpp":  "cpp",
		".java": "java",
		".rb":   "ruby",
		".sql":  "sql",
		".xml":  "xml",
	}
	return ftMap[ext]
}
