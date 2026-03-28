package picker

import (
	"context"
	"fmt"
	ext "nvim-gui/features/picker/external-tools"
	filewatcher "nvim-gui/features/picker/file-watcher"
	"nvim-gui/features/picker/highlighting"
	"nvim-gui/features/picker/match"
	"nvim-gui/utils"

	"github.com/charmbracelet/log"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type LiveGrepPicker struct {
	searchContext SearchContext
	Ctx           context.Context
	App           *application.App
	State         LiveGrepPickerState
}

type LiveGrepPickerState struct {
	SearchTerm     string     `json:"searchTerm"`
	Dir            string     `json:"dir"`
	Include        string     `json:"include"`
	Exclude        string     `json:"exclude"`
	CaseSensitive  bool       `json:"caseSensitive"`
	Regex          bool       `json:"regex"`
	MatchWholeWord bool       `json:"matchWholeWord"`
	Pagination     Pagination `json:"-"`
	TotalResults   int        `json:"totalResults"`
	TotalFiles     int        `json:"totalFiles"`
}

func NewLiveGrepPicker() *LiveGrepPicker {
	return &LiveGrepPicker{
		State: LiveGrepPickerState{
			SearchTerm: "",
			//TODO: this should be neovims cwd
			Dir:            "/home/stefan/Projects/nvim-gui",
			Include:        "",
			Exclude:        "",
			CaseSensitive:  false,
			Regex:          false,
			MatchWholeWord: false,
			Pagination: Pagination{
				PageIndex: 1,
				pages:     []Page{},
			},
			TotalResults: 0,
			TotalFiles:   0,
		},
	}
}

func (lgp *LiveGrepPicker) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	lgp.Ctx = ctx
	return nil
}

type SearchContext struct {
	ctx         context.Context
	cancel_func context.CancelFunc
}
type LiveGrepPage struct {
	GroupedMatches []match.MatchesOfFile `json:"results"`
	PageIndex      int                   `json:"pageIndex"`
	TotalPages     int                   `json:"totalPages"`
	TotalResults   int                   `json:"totalResults"`
	TotalFiles     int                   `json:"totalFiles"`
}

// ---------API START----------

func (lgp *LiveGrepPicker) GetLiveGrepOpts() LiveGrepPickerState {
	return lgp.State
}

func (p *LiveGrepPicker) LiveGrep(
	searchTerm string,
	dir string,
	exclude string,
	include string,
	case_sensitive bool,
	regex bool,
	match_whole_word bool,
) LiveGrepPage {
	if searchTerm == "" {
		return LiveGrepPage{}
	}
	if p.searchContext.cancel_func != nil {
		p.searchContext.cancel_func()
	}
	var searchContext context.Context
	searchContext, p.searchContext.cancel_func = context.WithCancel(context.Background())
	ctx, update_dir := filewatcher.InitContext(p.State.Dir, dir, p.Ctx)
	p.State.Dir = update_dir
	rg_lines, err := ext.Ripgrep(
		searchContext,
		searchTerm,
		"",
		dir,
		include,
		exclude,
		case_sensitive,
		regex,
		match_whole_word,
		false,
	)
	if err != nil {
		log.Debug(fmt.Sprintf("ripgrep error: %s", err))
		if ctx.Err() == context.Canceled {
			log.Debug("Search was canceled")
		}
		return LiveGrepPage{}
	}
	p.State.Pagination = MapPagination(rg_lines)
	matches := mapMatches(p.State.Pagination.pages[0].matches, searchTerm, regex)
	grouped_matches := match.MapSearchResult(matches)
	dirs := match.MapDirs(grouped_matches)
	go filewatcher.WatchFiles(ctx, dirs, dir, OnWrite, OnDelete)
	paths := utils.MapArray(rg_lines, func(line string) string {
		return ext.MapRipgrepInfo(line).Path
	})
	total_files := utils.CountUniqueItems(paths)
	return LiveGrepPage{
		GroupedMatches: grouped_matches,
		TotalPages:     len(p.State.Pagination.pages),
		PageIndex:      p.State.Pagination.PageIndex,
		TotalResults:   len(rg_lines),
		TotalFiles:     total_files,
	}
}

func (p *LiveGrepPicker) GetPrevPage() LiveGrepPage {
	if len(p.State.Pagination.pages) == 0 {
		return LiveGrepPage{}
	}
	p.State.Pagination.PageIndex--
	if p.State.Pagination.PageIndex < 0 {
		p.State.Pagination.PageIndex = len(p.State.Pagination.pages) - 1
	}
	page := p.State.Pagination.pages[p.State.Pagination.PageIndex]
	matches := mapMatches(page.matches, p.State.SearchTerm, p.State.Regex)
	grouped_matches := match.MapSearchResult(matches)

	return LiveGrepPage{
		GroupedMatches: grouped_matches,
		TotalPages:     len(p.State.Pagination.pages),
		PageIndex:      p.State.Pagination.PageIndex,
		TotalResults:   p.State.TotalResults,
		TotalFiles:     p.State.TotalFiles,
	}
}

func (p *LiveGrepPicker) GetNextPage() LiveGrepPage {
	if len(p.State.Pagination.pages) == 0 {
		return LiveGrepPage{}
	}
	p.State.Pagination.PageIndex++
	if p.State.Pagination.PageIndex >= len(p.State.Pagination.pages) {
		p.State.Pagination.PageIndex = 0
	}
	page := p.State.Pagination.pages[p.State.Pagination.PageIndex]
	matches := mapMatches(page.matches, p.State.SearchTerm, p.State.Regex)
	grouped_matches := match.MapSearchResult(matches)

	return LiveGrepPage{
		GroupedMatches: grouped_matches,
		TotalPages:     len(p.State.Pagination.pages),
		PageIndex:      p.State.Pagination.PageIndex,
		TotalResults:   p.State.TotalResults,
		TotalFiles:     p.State.TotalFiles,
	}
}

// ---------API END------------

func mapMatches(matches []PageMatch, searchTerm string, regex bool) []match.Match {
	return utils.MapArrayConcurrent(matches, func(page_match PageMatch) match.Match {
		line := page_match.rgLine
		rg_info := ext.MapRipgrepInfo(line)
		m := match.MapMatch(
			line,
			rg_info.Path,
			rg_info.MatchedText,
			rg_info.Row,
			rg_info.Col,
			searchTerm,
			"",
			false,
			regex,
		)
		html, _ := highlighting.Highlight(
			m.MatchedLine,
			rg_info.Path,
			m.MatchedText,
			m.ReplacementText,
		)
		m.Html = html
		return m
	})
}
