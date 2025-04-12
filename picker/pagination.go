package picker

import (
	ext "nvim-gui/picker/external-tools"
	"nvim-gui/picker/highlighting"
	"nvim-gui/picker/match"
	"nvim-gui/utils"
)

type PageMatch struct {
	rgLine string 
	match  *match.Match 
}

type Page struct {
	index   int 
	matches []PageMatch
}
type Pagination struct {
	PageIndex int 
	pages     []Page
}

func (p *LiveGrepPicker) GetPrevPage() SearchResult {
	if len(p.State.Pagination.pages) == 0 {
		return SearchResult{}
	}
	p.State.Pagination.PageIndex--
	if p.State.Pagination.PageIndex < 0 {
		p.State.Pagination.PageIndex = len(p.State.Pagination.pages) - 1
	}
	page := p.State.Pagination.pages[p.State.Pagination.PageIndex]
	matches := utils.MapArrayConcurrent(page.matches, func(page_match PageMatch) match.Match {
		line := page_match.rgLine
		rg_info := ext.MapRipgrepInfo(line)
		m := match.MapMatch(
			line,
			rg_info.Path,
			rg_info.MatchedText,
			rg_info.Row,
			rg_info.Col,
			p.State.SearchTerm,
			"",
			false,
			p.State.Regex,
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

	return SearchResult{
		GroupedMatches: grouped_matches,
		TotalPages:     len(p.State.Pagination.pages),
		PageIndex:      p.State.Pagination.PageIndex,
		TotalResults:   p.State.TotalResults,
		TotalFiles:     p.State.TotalFiles,
	}

}

func (p *LiveGrepPicker) GetNextPage() SearchResult {

	if len(p.State.Pagination.pages) == 0 {
		return SearchResult{}
	}
	p.State.Pagination.PageIndex++
	if p.State.Pagination.PageIndex >= len(p.State.Pagination.pages) {
		p.State.Pagination.PageIndex = 0
	}
	page := p.State.Pagination.pages[p.State.Pagination.PageIndex]
	matches := utils.MapArrayConcurrent(page.matches, func(page_match PageMatch) match.Match {
		line := page_match.rgLine
		rg_info := ext.MapRipgrepInfo(line)
		m := match.MapMatch(
			line,
			rg_info.Path,
			rg_info.MatchedText,
			rg_info.Row,
			rg_info.Col,
			p.State.SearchTerm,
			"",
			false,
			p.State.Regex,
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

	return SearchResult{
		GroupedMatches: grouped_matches,
		TotalPages:     len(p.State.Pagination.pages),
		PageIndex:      p.State.Pagination.PageIndex,
		TotalResults:   p.State.TotalResults,
		TotalFiles:     p.State.TotalFiles,
	}
}

func MapPagination(rg_lines []string) Pagination {
	chunks := utils.ChunkSlice(rg_lines, page_size)
	var pages []Page
	for i := 0; i < len(chunks); i++ {
		rg_lines := chunks[i]
		matches := utils.MapArray(rg_lines, func(line string) PageMatch {
			return PageMatch{
				rgLine: line,
				match:  nil,
			}
		})
		page := Page{
			index:   i,
			matches: matches,
		}
		pages = append(pages, page)
	}
	return Pagination{
		PageIndex: 0,
		pages:     pages,
	}
}
