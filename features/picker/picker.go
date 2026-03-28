package picker

import (
	"bytes"
	"context"
	"fmt"
	undo "nvim-gui/features/picker/undo"
	"nvim-gui/neovim"
	"nvim-gui/utils"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	INFO_LEVEL    = "info"
	SUCCESS_LEVEL = "success"
	WARNING_LEVEL = "warning"
	ERROR_LEVEL   = "error"
	DELETE        = "file-deleted"
	REPLACE       = "file-replaced"
	REPLACE_ALL   = "replaced-all"
	UNDO          = "undo"
	TOAST         = "toast"
	write_event   = REPLACE
	page_size     = 20
)

type PickerResult struct {
	Filename     string `json:"filename"`
	RelativePath string `json:"relativePath"`
	AbsolutePath string `json:"absolutePath"`
	Icon         string `json:"icon"`
	IconColor    string `json:"iconColor"`
	Text         string `json:"text"`
	Row          int    `json:"row"`
	Col          int    `json:"col"`
}

type Picker struct {
	undoStack undo.UndoStack
	Ctx       context.Context
	App       *application.App
	cwd       *string
}

func NewPicker() *Picker {
	return &Picker{
		undoStack: undo.UndoStack{},
	}
}

func (p *Picker) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	p.Ctx = ctx
	return nil
}

func (p *Picker) FindFiles(query string) []*PickerResult {
	if strings.TrimSpace(query) == "" {
		return []*PickerResult{}
	}
	if p.cwd == nil {
		cwd := neovim.GetCwd()
		p.cwd = &cwd
	}
	assert(p.cwd != nil, "cannot find git files without cwd")
	dir := utils.GetGitRepoRoot(*p.cwd)
	results, err := FindFiles(dir, query)
	if err != nil {
		log.Debug("FindFiles error getting results:\n", err.Error())
		return []*PickerResult{}
	}
	for _, result := range results {
		result.AbsolutePath = dir + "/" + result.RelativePath
	}
	return results
}

func (p *Picker) ClosePreview(winId int) {
	neovim.ClosePreview(winId)
}

func (p *Picker) GetPreview(filepath string, row int, col int) {
	neovim.ShowPreview(filepath, row, col)
	neovim.NvimScreen.EmitFloatingWindows()
}

func (p *Picker) OpenFile(path string, row int, col int) {
	log.Debug(fmt.Sprintf("open neovim file path=%s row=%d col=%d", path, row, col))
	err := neovim.OpenFileAt(path, row, col)
	if err != nil {
		log.Error(err.Error())
	}
	neovim.EmitEvent("hide-live-rep", struct{}{})
}

func assert(assertion bool, message string) {
	if !assertion {
		panic(message)
	}
}
func filterWithFzf(lines []string, query string) ([]string, error) {
	input := strings.Join(lines, "\n")
	fzf := exec.Command("fzf", "--filter="+query, "--delimiter=:", "--with-nth=1,2,3,4")
	fzf.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	fzf.Stdout = &out
	if err := fzf.Run(); err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(out.String()), "\n"), nil
}
