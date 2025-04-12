package picker

import (
	"context"
	"fmt"
	ext "nvim-gui/picker/external-tools"
	filewatcher "nvim-gui/picker/file-watcher"
	"nvim-gui/picker/highlighting"
	"nvim-gui/picker/match"
	"nvim-gui/utils"
	"time"
)

type LiveGrepPicker struct {
	searchContext SearchContext
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

type SearchContext struct {
	ctx         context.Context
	cancel_func context.CancelFunc
}
type SearchResult struct {
	GroupedMatches []match.MatchesOfFile
	PageIndex      int
	TotalPages     int
	TotalResults   int
	TotalFiles     int
}

func (p *LiveGrepPicker) SendKey(key string, ctrl bool, alt bool, shift bool) {
	//TODO: handle keymaps
}

func (p *LiveGrepPicker) LiveGrep(
	searchTerm string,
	dir string,
	exclude string,
	include string,
	case_sensitive bool,
	regex bool,
	match_whole_word bool,
	ctx context.Context,
) SearchResult {

	if searchTerm == "" {
		return SearchResult{}
	}
	utils.Log(fmt.Sprintf("LiveGrep searching for %s in dir %s", searchTerm, dir))
	utils.StartTime = time.Now()
	if p.searchContext.cancel_func != nil {
		p.searchContext.cancel_func()
	}
	var searchContext context.Context
	searchContext, p.searchContext.cancel_func = context.WithCancel(context.Background())
	ctx, update_dir := filewatcher.InitContext(p.State.Dir, dir, ctx)
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
		utils.Log(fmt.Sprintf("ripgrep error: %s", err))
		if ctx.Err() == context.Canceled {
			utils.Log("Search was canceled")
		}
		return SearchResult{}
	}
	p.State.Pagination = MapPagination(rg_lines)
	matches := utils.MapArrayConcurrent(p.State.Pagination.pages[0].matches, func(page_match PageMatch) match.Match {
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
	grouped_matches := match.MapSearchResult(matches)
	dirs := match.MapDirs(grouped_matches)
	// TODO: how to handle errors in a go routine?
	go filewatcher.WatchFiles(ctx, dirs, dir, OnWrite, OnDelete)

	paths := utils.MapArray(rg_lines, func(line string) string {
		return ext.MapRipgrepInfo(line).Path
	})
	total_files := utils.CountUniqueItems(paths)
	return SearchResult{
		GroupedMatches: grouped_matches,
		TotalPages:     len(p.State.Pagination.pages),
		PageIndex:      p.State.Pagination.PageIndex,
		TotalResults:   len(rg_lines),
		TotalFiles:     total_files,
	}
}
