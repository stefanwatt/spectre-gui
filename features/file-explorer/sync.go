package fileexplorer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	path "path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
)

type DirDraft struct {
	BufNr           int
	DirPath         string
	OriginalEntries map[uint64]DirectoryEntry
}

type SyncAction struct {
	Kind  string `json:"kind"`
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
	IsDir bool   `json:"isDir"`
}

type FileExplorerConfirmPrompt struct {
	Message string   `json:"message"`
	Choices []string `json:"choices"`
}

type syncIntent struct {
	ID      uint64
	From    string
	HasFrom bool
	To      string
	HasTo   bool
	IsDir   bool
}

func (e *FileExplorer) captureDirtyBaseline(bufNr int, directory *Directory) {
	if _, ok := e.dirtyByBuf[bufNr]; ok {
		return
	}
	originalEntries := make(map[uint64]DirectoryEntry, len(directory.Entries))
	for _, entry := range directory.Entries {
		originalEntries[entry.ID] = entry
	}
	e.dirtyByBuf[bufNr] = &DirDraft{
		BufNr:           bufNr,
		DirPath:         directory.Path,
		OriginalEntries: originalEntries,
	}
	log.Infof("[FileExplorer] captured dirty baseline buf=%d path=%s entries=%d", bufNr, directory.Path, len(originalEntries))
}

func (e *FileExplorer) indexSourceEntries(entries []DirectoryEntry) {
	for _, entry := range entries {
		if _, exists := e.sourceByID[entry.ID]; exists {
			continue
		}
		e.sourceByID[entry.ID] = entry
	}
}

func (e *FileExplorer) RequestClose() {
	if !e.active {
		return
	}
	if len(e.dirtyByBuf) == 0 {
		e.Close()
		return
	}
	e.showClosePrompt()
}

func (e *FileExplorer) HandleConfirmChoice(choice int) {
	if !e.pendingClosePrompt {
		return
	}
	e.hideClosePrompt()
	switch choice {
	case 1: // Sync
		if err := e.Sync(); err != nil {
			log.Errorf("[FileExplorerClose] sync on close failed: %v", err)
			return
		}
		e.Close()
	case 2: // Discard
		e.dirtyByBuf = map[int]*DirDraft{}
		e.Close()
	default: // Cancel
		return
	}
}

func (e *FileExplorer) showClosePrompt() {
	if e.emitter == nil {
		log.Warn("[FileExplorerClose] no emitter configured, keeping explorer open")
		return
	}
	e.pendingClosePrompt = true
	e.emitter.Emit("file-explorer-confirm-prompt-show", FileExplorerConfirmPrompt{
		Message: "Unsynced file-explorer changes.\nChoose action:",
		Choices: []string{"Sync", "Discard", "Cancel"},
	})
}

func (e *FileExplorer) hideClosePrompt() {
	if !e.pendingClosePrompt {
		return
	}
	e.pendingClosePrompt = false
	if e.emitter != nil {
		e.emitter.Emit("file-explorer-confirm-prompt-hide", struct{}{})
	}
}

func (e *FileExplorer) BuildSyncActionPlan() ([]SyncAction, error) {
	if len(e.dirtyByBuf) == 0 {
		return []SyncAction{}, nil
	}

	bufNrs := make([]int, 0, len(e.dirtyByBuf))
	for bufNr := range e.dirtyByBuf {
		bufNrs = append(bufNrs, bufNr)
	}
	sort.Ints(bufNrs)

	allIntents := make([]syncIntent, 0, 32)
	for _, bufNr := range bufNrs {
		draft := e.dirtyByBuf[bufNr]
		intents, err := e.collectDraftIntents(draft)
		if err != nil {
			return nil, err
		}
		allIntents = append(allIntents, intents...)
	}

	actions := classifySyncActions(allIntents)
	return actions, nil
}

func (e *FileExplorer) Sync() error {
	actions, err := e.BuildSyncActionPlan()
	if err != nil {
		return err
	}

	e.logSyncActionPlan(actions)
	if err := e.applySyncActions(actions); err != nil {
		e.echoSyncMessage(fmt.Sprintf("file-explorer sync failed: %v", err))
		return err
	}

	e.dirtyByBuf = map[int]*DirDraft{}
	if err := e.refreshDirectoriesFromDisk(); err != nil {
		e.echoSyncMessage(fmt.Sprintf("file-explorer sync failed: %v", err))
		return err
	}
	e.clearPreview()
	if err := e.refreshVisiblePanes(); err != nil {
		e.echoSyncMessage(fmt.Sprintf("file-explorer sync failed: %v", err))
		return err
	}
	e.Dirty = true
	e.echoSyncMessage("file-explorer sync: success")
	return nil
}

func (e *FileExplorer) refreshDirectoriesFromDisk() error {
	paths := make([]string, 0, len(e.directoriesByPath))
	for dirPath := range e.directoriesByPath {
		paths = append(paths, dirPath)
	}
	sort.Strings(paths)

	for _, dirPath := range paths {
		directory := e.directoriesByPath[dirPath]
		selectedPath := selectedEntryPath(directory)
		if err := e.loadDirectory(directory, dirPath); err != nil {
			if os.IsNotExist(err) {
				directory.Entries = []DirectoryEntry{}
				directory.SelectedEntryId = 0
				if err := e.setBufferLinesFromEntries(directory.BufNr, directory.Entries); err != nil {
					return err
				}
				continue
			}
			return err
		}
		restoreSelectedPath(directory, selectedPath)
		if err := e.setBufferLinesFromEntries(directory.BufNr, directory.Entries); err != nil {
			return err
		}
	}
	return nil
}

func selectedEntryPath(directory *Directory) string {
	if directory == nil {
		return ""
	}
	for _, entry := range directory.Entries {
		if entry.ID == directory.SelectedEntryId {
			return entry.Path
		}
	}
	return ""
}

func restoreSelectedPath(directory *Directory, selectedPath string) {
	if directory == nil {
		return
	}
	if selectedPath == "" {
		if len(directory.Entries) > 0 {
			directory.SelectedEntryId = directory.Entries[0].ID
		} else {
			directory.SelectedEntryId = 0
		}
		return
	}

	for _, entry := range directory.Entries {
		if entry.Path == selectedPath {
			directory.SelectedEntryId = entry.ID
			return
		}
	}
	if len(directory.Entries) > 0 {
		directory.SelectedEntryId = directory.Entries[0].ID
	} else {
		directory.SelectedEntryId = 0
	}
}

func (e *FileExplorer) applySyncActions(actions []SyncAction) error {
	if len(actions) == 0 {
		return nil
	}

	ordered := sortActionsForApply(actions)
	for _, action := range ordered {
		if err := e.applySyncAction(action); err != nil {
			return fmt.Errorf("%s failed (%s): %w", action.Kind, formatAction(action), err)
		}
	}
	return nil
}

func sortActionsForApply(actions []SyncAction) []SyncAction {
	ordered := append([]SyncAction(nil), actions...)
	rank := map[string]int{
		"copy":   0,
		"create": 1,
		"move":   2,
		"rename": 3,
		"delete": 4,
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		ri := rank[ordered[i].Kind]
		rj := rank[ordered[j].Kind]
		if ri != rj {
			return ri < rj
		}
		if ordered[i].From == ordered[j].From {
			return ordered[i].To < ordered[j].To
		}
		return ordered[i].From < ordered[j].From
	})
	return ordered
}

func (e *FileExplorer) applySyncAction(action SyncAction) error {
	switch action.Kind {
	case "copy":
		if action.From == "" || action.To == "" {
			return fmt.Errorf("missing from/to")
		}
		if action.IsDir {
			return copyDirectory(action.From, action.To)
		}
		return copyFile(action.From, action.To)
	case "create":
		if action.To == "" {
			return fmt.Errorf("missing to")
		}
		if action.IsDir {
			return createDirectory(action.To)
		}
		return createFile(action.To)
	case "move", "rename":
		if action.From == "" || action.To == "" {
			return fmt.Errorf("missing from/to")
		}
		if err := ensureParentDir(action.To); err != nil {
			return err
		}
		if err := ensurePathAbsent(action.To); err != nil {
			return err
		}
		return os.Rename(action.From, action.To)
	case "delete":
		if action.From == "" {
			return fmt.Errorf("missing from")
		}
		return os.RemoveAll(action.From)
	default:
		return fmt.Errorf("unknown action kind %q", action.Kind)
	}
}

func createDirectory(target string) error {
	if err := ensurePathAbsent(target); err != nil {
		return err
	}
	return os.MkdirAll(target, 0o755)
}

func createFile(target string) error {
	if err := ensureParentDir(target); err != nil {
		return err
	}
	if err := ensurePathAbsent(target); err != nil {
		return err
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}

func ensureParentDir(target string) error {
	parent := path.Dir(target)
	if parent == "." || parent == "" {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}

func ensurePathAbsent(target string) error {
	_, err := os.Stat(target)
	if err == nil {
		return fmt.Errorf("target already exists: %s", target)
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func copyFile(from, to string) error {
	info, err := os.Stat(from)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("source is directory: %s", from)
	}
	if err := ensureParentDir(to); err != nil {
		return err
	}
	if err := ensurePathAbsent(to); err != nil {
		return err
	}

	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func copyDirectory(from, to string) error {
	rootInfo, err := os.Stat(from)
	if err != nil {
		return err
	}
	if !rootInfo.IsDir() {
		return fmt.Errorf("source is not directory: %s", from)
	}
	if err := ensureParentDir(to); err != nil {
		return err
	}
	if err := ensurePathAbsent(to); err != nil {
		return err
	}

	return path.WalkDir(from, func(current string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := path.Rel(from, current)
		if err != nil {
			return err
		}
		targetPath := to
		if rel != "." {
			targetPath = path.Join(to, rel)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}
		if err := ensureParentDir(targetPath); err != nil {
			return err
		}
		return copyFile(current, targetPath)
	})
}

func formatAction(action SyncAction) string {
	if action.From != "" && action.To != "" {
		return action.From + " -> " + action.To
	}
	if action.To != "" {
		return action.To
	}
	return action.From
}

func (e *FileExplorer) collectDraftIntents(draft *DirDraft) ([]syncIntent, error) {
	lineBytes, err := e.nvim.GetBufferLines(draft.BufNr, 0, -1, false)
	if err != nil {
		return nil, fmt.Errorf("get buffer lines for buf=%d failed: %w", draft.BufNr, err)
	}

	intents := make([]syncIntent, 0, len(lineBytes)+len(draft.OriginalEntries))
	parsedIDs := make(map[uint64]bool, len(lineBytes))
	currentPathsByID := make(map[uint64]map[string]struct{}, len(lineBytes))

	for _, rawLine := range lineBytes {
		line := strings.TrimSpace(string(rawLine))
		if line == "" {
			continue
		}
		id, text := parseBufferLine(line)
		name, isDir := parseDraftTarget(text)
		if name == "" {
			continue
		}

		toPath := path.Join(draft.DirPath, name)
		intent := syncIntent{To: toPath, HasTo: true, ID: id, IsDir: isDir}

		if id != 0 {
			parsedIDs[id] = true
			if currentPathsByID[id] == nil {
				currentPathsByID[id] = map[string]struct{}{}
			}
			currentPathsByID[id][toPath] = struct{}{}
			if sourceEntry, ok := e.sourceByID[id]; ok {
				intent.From = sourceEntry.Path
				intent.HasFrom = true
				if sourceEntry.IsDir {
					intent.IsDir = true
				}
			}
		}

		intents = append(intents, intent)
	}

	for id, original := range draft.OriginalEntries {
		if !parsedIDs[id] {
			intents = append(intents, syncIntent{
				ID:      id,
				From:    original.Path,
				HasFrom: true,
				IsDir:   original.IsDir,
			})
			continue
		}

		currentPaths := currentPathsByID[id]
		if _, stillAtOriginalPath := currentPaths[original.Path]; stillAtOriginalPath {
			continue
		}
		intents = append(intents, syncIntent{
			ID:      id,
			From:    original.Path,
			HasFrom: true,
			IsDir:   original.IsDir,
		})
	}

	return intents, nil
}

func parseDraftTarget(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}
	isDir := strings.HasSuffix(trimmed, "/")
	name := strings.TrimSuffix(trimmed, "/")
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	return name, isDir
}

func classifySyncActions(intents []syncIntent) []SyncAction {
	creates := make([]SyncAction, 0)
	deletes := make([]SyncAction, 0)
	rawCopies := make([]SyncAction, 0)

	for _, intent := range intents {
		switch {
		case !intent.HasFrom && intent.HasTo:
			creates = append(creates, SyncAction{Kind: "create", To: intent.To, IsDir: intent.IsDir})
		case intent.HasFrom && !intent.HasTo:
			deletes = append(deletes, SyncAction{Kind: "delete", From: intent.From, IsDir: intent.IsDir})
		case intent.HasFrom && intent.HasTo:
			if intent.From == intent.To {
				continue
			}
			rawCopies = append(rawCopies, SyncAction{Kind: "copy", From: intent.From, To: intent.To, IsDir: intent.IsDir})
		}
	}

	sort.Slice(creates, func(i, j int) bool { return creates[i].To < creates[j].To })
	sort.Slice(deletes, func(i, j int) bool { return deletes[i].From < deletes[j].From })
	sort.Slice(rawCopies, func(i, j int) bool {
		if rawCopies[i].From == rawCopies[j].From {
			return rawCopies[i].To < rawCopies[j].To
		}
		return rawCopies[i].From < rawCopies[j].From
	})

	copiesByFrom := map[string][]SyncAction{}
	for _, copyAction := range rawCopies {
		copiesByFrom[copyAction.From] = append(copiesByFrom[copyAction.From], copyAction)
	}
	for from := range copiesByFrom {
		sort.Slice(copiesByFrom[from], func(i, j int) bool { return copiesByFrom[from][i].To < copiesByFrom[from][j].To })
	}

	deletesByFrom := map[string][]SyncAction{}
	for _, deleteAction := range deletes {
		deletesByFrom[deleteAction.From] = append(deletesByFrom[deleteAction.From], deleteAction)
	}

	deleteFromPaths := make([]string, 0, len(deletesByFrom))
	for from := range deletesByFrom {
		deleteFromPaths = append(deleteFromPaths, from)
	}
	sort.Strings(deleteFromPaths)

	moves := make([]SyncAction, 0)
	renames := make([]SyncAction, 0)
	remainingDeletes := make([]SyncAction, 0)

	for _, from := range deleteFromPaths {
		deleteList := deletesByFrom[from]
		copyList := copiesByFrom[from]
		for _, deleteAction := range deleteList {
			if len(copyList) == 0 {
				remainingDeletes = append(remainingDeletes, deleteAction)
				continue
			}
			collapsedCopy := copyList[0]
			copyList = copyList[1:]
			collapsed := SyncAction{
				From:  from,
				To:    collapsedCopy.To,
				IsDir: deleteAction.IsDir || collapsedCopy.IsDir,
			}
			if path.Dir(from) == path.Dir(collapsedCopy.To) {
				collapsed.Kind = "rename"
				renames = append(renames, collapsed)
			} else {
				collapsed.Kind = "move"
				moves = append(moves, collapsed)
			}
		}
		copiesByFrom[from] = copyList
	}

	remainingCopies := make([]SyncAction, 0)
	copyFromPaths := make([]string, 0, len(copiesByFrom))
	for from := range copiesByFrom {
		copyFromPaths = append(copyFromPaths, from)
	}
	sort.Strings(copyFromPaths)
	for _, from := range copyFromPaths {
		remainingCopies = append(remainingCopies, copiesByFrom[from]...)
	}

	sort.Slice(moves, func(i, j int) bool {
		if moves[i].From == moves[j].From {
			return moves[i].To < moves[j].To
		}
		return moves[i].From < moves[j].From
	})
	sort.Slice(renames, func(i, j int) bool {
		if renames[i].From == renames[j].From {
			return renames[i].To < renames[j].To
		}
		return renames[i].From < renames[j].From
	})
	sort.Slice(remainingDeletes, func(i, j int) bool { return remainingDeletes[i].From < remainingDeletes[j].From })

	actions := make([]SyncAction, 0, len(remainingCopies)+len(creates)+len(moves)+len(renames)+len(remainingDeletes))
	actions = append(actions, remainingCopies...)
	actions = append(actions, creates...)
	actions = append(actions, moves...)
	actions = append(actions, renames...)
	actions = append(actions, remainingDeletes...)

	return dedupeSyncActions(actions)
}

func dedupeSyncActions(actions []SyncAction) []SyncAction {
	seen := map[string]struct{}{}
	result := make([]SyncAction, 0, len(actions))
	for _, action := range actions {
		key := fmt.Sprintf("%s|%s|%s|%t", action.Kind, action.From, action.To, action.IsDir)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, action)
	}
	return result
}

func (e *FileExplorer) logSyncActionPlan(actions []SyncAction) {
	if len(actions) == 0 {
		log.Info("[FileExplorerSync] no pending actions")
		return
	}

	payload, err := json.MarshalIndent(actions, "", "  ")
	if err != nil {
		log.Errorf("[FileExplorerSync] marshal action plan failed: %v", err)
		return
	}
	log.Infof("[FileExplorerSync] planned actions:\n%s", string(payload))
}

func (e *FileExplorer) echoSyncMessage(message string) {
	escaped := strings.ReplaceAll(message, "'", "''")
	if err := e.nvim.Command(fmt.Sprintf("echo '%s'", escaped)); err != nil {
		log.Errorf("[FileExplorerSync] echo failed: %v", err)
	}
}
