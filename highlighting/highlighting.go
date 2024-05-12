package highlighting

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"spectre-gui/utils"

	html_formatter "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func Highlight(code string, filename string, col int, matched_text string, replacement string) (string, string) {
	utils.Log(fmt.Sprintf("Highlighting code \n'%s' \nwith replacement: %s\non file%s", code, replacement, filename))
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Get("plaintext")
		if lexer == nil {
			panic("Plaintext lexer not available")
		}
	}
	var bytes bytes.Buffer
	formatter := html_formatter.New(
		html_formatter.WithClasses(true),
		html_formatter.ClassPrefix("spectre-"),
		html_formatter.InlineCode(true),
	)
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		log.Fatal(err)
	}
	style := styles.Get("catppuccin-frappe")
	if style == nil {
		style = styles.Fallback
	}
	err = formatter.Format(&bytes, style, iterator)
	if err != nil {
		log.Fatal(err)
	}
	highlighted := bytes.String()
	utils.Log(fmt.Sprintf("highlighted code: %s", highlighted))
	html := inject_match(highlighted, matched_text, replacement)
	formatter.WriteCSS(&bytes, style)
	return html, bytes.String()
}

func inject_match(html string, matched_text string, replacement string) string {
	if matched_text == "" {
		utils.Log("matched text empty -> cannot inject match html")
		return html
	}
	replacement_html := ""
	if replacement != "" {
		replacement_html = fmt.Sprintf("<span class=\"spectre-replacement\">%s</span>", replacement)
	}
	match_html := fmt.Sprintf("<span class=\"spectre-matched\">%s</span>%s", matched_text, replacement_html)
	children, err := get_children(html)

	if err != nil {
		utils.Log(err.Error())
	} else {
		utils.Log("children:", children)
	}
	before, middle, after := split_spans(matched_text, children)
	utils.Log("inject match ; matched text: " + matched_text)
	utils.Log("inject match;before:", before)
	utils.Log("inject match;middle:", middle)
	utils.Log("inject match;after:", after)
	before_html := concat_outer_html(before)
	middle_html := get_middle_html(middle, matched_text, match_html)
	utils.Log(fmt.Sprintf("middle html: %s", middle_html))
	after_html := concat_outer_html(after)

	return fmt.Sprintf(
		`<code class="spectre-chroma">%s%s%s</code`,
		before_html,
		middle_html,
		after_html,
	)
}

func get_children(html_content string) ([]ChromaSpan, error) {
	doc, err := html.Parse(strings.NewReader(html_content))
	if err != nil {
		return nil, err
	}

	var spans []ChromaSpan
	var traverse func(*html.Node, bool)
	traverse = func(node *html.Node, isSpan bool) {
		if node.Type == html.ElementNode && node.Data == "span" {
			var outer_html, innerHtml bytes.Buffer
			html.Render(&outer_html, node)
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				html.Render(&innerHtml, c)
			}
			decoded_inner_html := html.UnescapeString(innerHtml.String())
			outer_html_str := outer_html.String()
			whitespace_suffix := ""
			if next_sibling := node.NextSibling; next_sibling != nil && next_sibling.Type == html.TextNode {
				whitespace := next_sibling.Data
				if whitespace_index := strings.IndexFunc(whitespace, func(r rune) bool {
					return !unicode.IsSpace(r)
				}); whitespace_index == -1 {
					whitespace_suffix = whitespace
				} else if whitespace_index > 0 {
					whitespace_suffix = whitespace[:whitespace_index]
				}
			}
			outer_html_str += whitespace_suffix
			decoded_inner_html += whitespace_suffix
			spans = append(spans, ChromaSpan{OuterHtml: outer_html_str, InnerHtml: decoded_inner_html})
		} else if node.Type == html.TextNode && !isSpan {
			if len(spans) > 0 && strings.TrimSpace(node.Data) != "" {
				spans[len(spans)-1].OuterHtml += node.Data
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c, node.Type == html.ElementNode && node.Data == "span")
		}
	}

	traverse(doc, false)
	return spans, nil
}

func get_middle_html(spans []ChromaSpan, matched_text string, match_html string) string {
	if len(spans) == 0 {
		return match_html
	}

	var result bytes.Buffer
	concatenated_html := concat_inner_html(spans)
	start_match := strings.Index(concatenated_html, matched_text)

	if start_match == -1 {
		return ""
	}
	end_match := start_match + len(matched_text)
	current_html_index := 0
	for _, span := range spans {
		span_length := len(span.InnerHtml)
		span_end_html_index := current_html_index + span_length
		if start_match >= current_html_index && start_match < span_end_html_index {
			relative_start := start_match - current_html_index
			relative_end := end_match - current_html_index
			if relative_start < 0 {
				relative_start = 0
			}
			if relative_end > span_length {
				relative_end = span_length
			}
			if relative_start > 0 && relative_start <= span_length {
				result.WriteString(span.OuterHtml[:strings.Index(span.OuterHtml, span.InnerHtml)+relative_start])
			}
			result.WriteString(match_html)
			if relative_end < span_length && relative_end >= 0 {
				after_match_start := strings.LastIndex(span.OuterHtml, span.InnerHtml) + relative_end
				if after_match_start < len(span.OuterHtml) {
					result.WriteString(span.OuterHtml[after_match_start:])
				}
			}
			start_match = int(^uint(0) >> 1)
		} else {
			if start_match > span_end_html_index {
				result.WriteString(span.OuterHtml)
			}
		}
		current_html_index = span_end_html_index
	}
	return result.String()
}

func concat_inner_html(spans []ChromaSpan) string {
	var result bytes.Buffer
	for _, span := range spans {
		result.WriteString(span.InnerHtml)
	}
	return result.String()
}

func concat_outer_html(spans []ChromaSpan) string {
	outer_htmls := utils.MapArray(spans, func(span ChromaSpan) string {
		return fmt.Sprintf(span.OuterHtml)
	})
	return strings.Join(outer_htmls, "")
}

type ChromaSpan struct {
	OuterHtml string
	InnerHtml string
}

func split_spans(matched_text string, children []ChromaSpan) ([]ChromaSpan, []ChromaSpan, []ChromaSpan) {
	if len(matched_text) == 0 {
		return nil, nil, children
	}

	var before, middle, after []ChromaSpan
	startIndex, endIndex := -1, -1
	current_text := ""
	indexMap := make(map[int]int)
	for i, child := range children {
		current_text += child.InnerHtml
		for j := 0; j < len(child.InnerHtml); j++ {
			indexMap[len(current_text)-len(child.InnerHtml)+j] = i
		}
	}
	utils.Log(fmt.Sprintf("index of %s in combined innerHTML %s is", matched_text, current_text))
	startPos := strings.Index(current_text, matched_text)
	if startPos == -1 {
		return children, nil, nil
	}
	utils.Log(fmt.Sprintf("index = %d", startPos))
	endPos := startPos + len(matched_text) - 1
	startIndex = indexMap[startPos]
	endIndex = indexMap[endPos]
	if startIndex > 0 {
		before = children[:startIndex]
	}
	if startIndex != -1 && endIndex != -1 && endIndex >= startIndex {
		middle = children[startIndex : endIndex+1]
	}
	if endIndex+1 < len(children) {
		after = children[endIndex+1:]
	}

	return before, middle, after
}

func add_delimiter(code string, col int, matched_text string) (string, string) {
	before := code[:col-1]
	middle := code[col : col+len(matched_text)]
	after := code[col+len(matched_text)-1:]
	delimiter, err := utils.RandomString(6)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s%s%s%s%s", before, delimiter, middle, delimiter, after), delimiter
}
