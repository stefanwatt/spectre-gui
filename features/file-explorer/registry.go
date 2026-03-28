package fileexplorer

import "sync"

type PaneAssignment struct {
	WinID int
	BufNr int
}
type Registry struct {
	mu                 sync.RWMutex
	active             bool
	parent             PaneAssignment
	current            PaneAssignment
	preview            PaneAssignment
	directoryLineIsDir map[int]map[int]bool // bufNr -> lineNr -> isDir
}
type RegistrySnapshot struct {
	Active             bool
	Parent             PaneAssignment
	Current            PaneAssignment
	Preview            PaneAssignment
	DirectoryLineIsDir map[int]map[int]bool
}

func NewRegistry() *Registry {
	return &Registry{
		directoryLineIsDir: make(map[int]map[int]bool),
		parent:             PaneAssignment{WinID: -1, BufNr: -1},
		current:            PaneAssignment{WinID: -1, BufNr: -1},
		preview:            PaneAssignment{WinID: -1, BufNr: -1},
	}
}

func (r *Registry) SetActive(active bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.active = active
}

func (r *Registry) UpdateParentPaneBufnr(bufnr int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// TODO: there is a bug: this is emitted twice when opening and only the first one is the correct parent bufnr
	if r.parent.BufNr == -1 {
		r.parent.BufNr = bufnr
	}
}

func (r *Registry) UpdateCurrentPaneBufnr(bufnr int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.current.BufNr = bufnr
}

func (r *Registry) UpdatePreviewPaneBufnr(bufnr int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preview.BufNr = bufnr
}

func (r *Registry) AssignParentPane(winId, bufnr int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.parent = PaneAssignment{WinID: winId, BufNr: bufnr}
}

func (r *Registry) AssignCurrentPane(winId, bufnr int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.current = PaneAssignment{WinID: winId, BufNr: bufnr}
}

func (r *Registry) AssignPreviewPane(winId, bufnr int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preview = PaneAssignment{WinID: winId, BufNr: bufnr}
}

func (r *Registry) SetDirectoryLineMap(bufnr int, lineMap map[int]bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.directoryLineIsDir[bufnr] = lineMap
}

func (r *Registry) Snapshot() RegistrySnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Deep copy directoryLineIsDir to avoid races
	dirCopy := make(map[int]map[int]bool, len(r.directoryLineIsDir))
	for bufNr, lineMap := range r.directoryLineIsDir {
		lmCopy := make(map[int]bool, len(lineMap))
		for k, v := range lineMap {
			lmCopy[k] = v
		}
		dirCopy[bufNr] = lmCopy
	}

	return RegistrySnapshot{
		Active:             r.active,
		Parent:             r.parent,
		Current:            r.current,
		Preview:            r.preview,
		DirectoryLineIsDir: dirCopy,
	}
}
