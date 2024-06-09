package neovim

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/google/uuid"

	"spectre-gui/utils"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type HighlightToken struct {
	Id            string `msgpack:"id" json:"id"`
	Text          string `msgpack:"text" json:"text"`
	StartRow      uint64 `msgpack:"start_row" json:"start_row"`
	EndRow        uint64 `msgpack:"end_row" json:"end_row"`
	StartCol      uint64 `msgpack:"start_col" json:"start_col"`
	EndCol        uint64 `msgpack:"end_col" json:"end_col"`
	Foreground    string `msgpack:"fg" json:"foreground"`
	Background    string `msgpack:"bg" json:"background"`
	Reverse       bool   `msgpack:"reverse" json:"reverse"`
	Underline     bool   `msgpack:"underline" json:"underline"`
	Undercurl     bool   `msgpack:"undercurl" json:"undercurl"`
	Strikethrough bool   `msgpack:"strikethrough" json:"strikethrough"`
	Bold          bool   `msgpack:"bold" json:"bold"`
	Italic        bool   `msgpack:"italic" json:"italic"`
	HlGroup       string `msgpack:"hl_group" json:"hl_group"`
}

type BufLine struct {
	Sign   string           `msgpack:"sign" json:"sign"`
	Row    uint64           `msgpack:"row" json:"row"`
	Tokens []HighlightToken `msgpack:"tokens" json:"tokens"`
}

type BufChangeEvent struct {
	Event             string
	Buffer            int64
	ChangedTick       int64
	FirstLine         int64
	LastLine          int64
	LastLineChanged   int64
	PreviousByteCount int64
	DeletedCodepoints *int64
	DeletedCodeunits  *int64
}

func GetBufLines(hl_tokens []HighlightToken) []BufLine {
	if len(hl_tokens) == 0 {
		return []BufLine{
			{
				Sign: "",
				Row:  uint64(0),
				Tokens: []HighlightToken{
					{
						Text:     " ",
						StartRow: uint64(0),
						StartCol: 0,
						EndCol:   1,
						EndRow:   uint64(1),
						Id:       uuid.NewString(),
					},
				},
			},
		}
	}
	hl_tokens = merge_tokens(hl_tokens)
	log.Println("num hl tokens: ", len(hl_tokens))
	slices.SortFunc(hl_tokens, func(a HighlightToken, b HighlightToken) int {
		if a.StartRow == b.StartRow {
			return int(a.StartCol - b.StartCol)
		} else {
			return int(a.StartRow - b.StartRow)
		}
	})
	buf_lines := map_buf_lines(hl_tokens)
	return buf_lines
}

func OnBufChanged(ctx context.Context, hl_tokens []HighlightToken) {
	if len(hl_tokens) < 50 {
		return
	}
	updated_buf_lines := GetBufLines(hl_tokens)
	utils.LogTimeSinceLast("update buf lines")
	Runtime.EventsEmit(ctx, "buf-lines-changed", updated_buf_lines)
}

func merge_tokens(tokens []HighlightToken) []HighlightToken {
	token_ranges := make(map[string][]HighlightToken)
	for _, token := range tokens {
		token_key := fmt.Sprintf("%d-%d-%d-%d", token.StartRow, token.StartCol, token.EndRow, token.EndCol)
		token.HlGroup = strings.ReplaceAll(token.HlGroup, ".", "-")
		token_ranges[token_key] = append(token_ranges[token_key], token)
	}

	merged_tokens := []HighlightToken{}

	for _, tokens_of_range := range token_ranges {
		if len(tokens_of_range) == 0 {
			continue
		}
		if len(tokens_of_range) == 1 {
			merged_tokens = append(merged_tokens, tokens_of_range[0])
		} else {
			merged_token := tokens_of_range[0]
			for i := 1; i < len(tokens_of_range); i++ {
				if tokens_of_range[i].Foreground != "" {
					merged_token.Foreground = tokens_of_range[i].Foreground
				}
				if tokens_of_range[i].Background != "" {
					merged_token.Background = tokens_of_range[i].Background
				}
				if tokens_of_range[i].HlGroup != "" {
					merged_token.HlGroup = tokens_of_range[i].HlGroup
				}
			}
			merged_tokens = append(merged_tokens, merged_token)
		}
	}
	return merged_tokens
}

func map_empty_buf_line(row int) BufLine {
	return BufLine{
		Sign: "",
		Row:  uint64(row),
		Tokens: []HighlightToken{
			{
				Text:     " ",
				StartRow: uint64(row),
				EndRow:   uint64(row),
				StartCol: 0,
				EndCol:   1,
				Id:       uuid.NewString(),
			},
		},
	}
}

func map_post_space_token(last_token_of_line HighlightToken, char_diff int) HighlightToken {
	return HighlightToken{
		Text:     strings.Repeat(" ", char_diff),
		StartRow: last_token_of_line.StartRow,
		EndRow:   last_token_of_line.EndRow,
		StartCol: last_token_of_line.StartCol - uint64(char_diff),
		EndCol:   last_token_of_line.EndCol - uint64(char_diff),
		Id:       uuid.NewString(),
	}
}

func map_indentation_token(first_token_of_line HighlightToken, start_col uint64) HighlightToken {
	return HighlightToken{
		Text:     strings.Repeat("\t", int(start_col)),
		StartRow: first_token_of_line.StartRow,
		EndRow:   first_token_of_line.EndRow,
		StartCol: first_token_of_line.StartCol - start_col,
		EndCol:   first_token_of_line.EndCol - start_col,
		Id:       uuid.NewString(),
	}
}

func map_buf_lines(tokens []HighlightToken) []BufLine {
	var buf_lines []BufLine
	row := 0
	var tokens_of_line []HighlightToken
	last_end_col := -1
	for _, token := range tokens {
		if row != int(token.StartRow) {
			buf_lines = append(buf_lines, BufLine{
				Sign:   "",
				Row:    uint64(row),
				Tokens: tokens_of_line,
			})
			tokens_of_line = []HighlightToken{}
			last_end_col = -1
			row++
		}
		for row < int(token.StartRow) {
			buf_lines = append(buf_lines, map_empty_buf_line(row))
			tokens_of_line = []HighlightToken{}
			last_end_col = -1
			row++
		}
		if last_end_col != -1 {
			char_diff := int(token.StartCol) - last_end_col
			if char_diff > 0 {
				tokens_of_line = append(tokens_of_line, map_post_space_token(token, char_diff))
			}
		} else if int(token.StartCol) > 0 {
			start_col := token.StartCol
			tokens_of_line = append(tokens_of_line, map_indentation_token(token, start_col))
		}
		last_end_col = int(token.EndCol)
		tokens_of_line = append(tokens_of_line, token)
	}
	return buf_lines
}
