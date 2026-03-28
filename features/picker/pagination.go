package picker

import (
	"nvim-gui/features/picker/match"
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
