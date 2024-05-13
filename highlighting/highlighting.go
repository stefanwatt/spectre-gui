package highlighting

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"spectre-gui/utils"

	"github.com/alecthomas/chroma/v2"
	html_formatter "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func Highlight(code string, lexer chroma.Lexer, filename string, col int, matched_text string, replacement string) (string, string) {
	utils.Log2(fmt.Sprintf("Highlighting code \n'%s' \nwith replacement: %s\non file%s", code, replacement, filename))
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
	utils.Log2(fmt.Sprintf("highlighted code: %s", highlighted))
	html := inject_match(highlighted, matched_text, replacement)
	formatter.WriteCSS(&bytes, style)
	return html, bytes.String()
}

var lexerCache = make(map[string]chroma.Lexer)

func MatchLexer(filename string) chroma.Lexer {
	ext := strings.TrimLeft(filepath.Ext(filename), ".")
	if cached_lexer, exists := lexerCache[ext]; exists {
		return cached_lexer
	} else {
		lexer := lexers.Get(ext)
		if lexer == nil {
			lexer = lexers.Fallback
		}
		lexerCache[ext] = lexer
		return lexer
	}
}

func inject_match(html string, matched_text string, replacement string) string {
	if matched_text == "" {
		utils.Log2("matched text empty -> cannot inject match html")
		return html
	}
	replacement_html := ""
	if replacement != "" {
		replacement_html = fmt.Sprintf("<span class=\"spectre-replacement\">%s</span>", replacement)
	}
	match_html := fmt.Sprintf("<span class=\"spectre-matched\">%s</span>%s", matched_text, replacement_html)
	children, err := get_children(html)

	if err != nil {
		utils.Log2(err.Error())
	} else {
		utils.Log2("children:", children)
	}
	before, middle, after := split_spans(matched_text, children)
	utils.Log2("inject match ; matched text: " + matched_text)
	utils.Log2("inject match;before:", before)
	utils.Log2("inject match;middle:", middle)
	utils.Log2("inject match;after:", after)
	before_html := concat_outer_html(before)
	utils.Log2(fmt.Sprintf("before html: %s", before_html))
	middle_html := get_middle_html(middle, matched_text, match_html)
	utils.Log2(fmt.Sprintf("middle html: %s", middle_html))
	after_html := concat_outer_html(after)
	utils.Log2(fmt.Sprintf("after html: %s", after_html))

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

func get_middle_html(containing_spans []ChromaSpan, matched_text string, match_html string) string {
	if len(containing_spans) == 0 {
		return match_html
	}

	var concatenated_inner_html bytes.Buffer
	for _, span := range containing_spans {
		concatenated_inner_html.WriteString(span.InnerHtml)
	}
	concatenatedHTML := concatenated_inner_html.String()

	startMatch := strings.Index(concatenatedHTML, matched_text)
	if startMatch == -1 {
		return ""
	}

	return inject_match_html(containing_spans, startMatch, len(matched_text), match_html)
}

func inject_match_html(containing_spans []ChromaSpan, match_start_index int, match_length int, match_html string) string {
	var result bytes.Buffer
	current_html_index := 0

	for _, span := range containing_spans {
		span_length := len(span.InnerHtml)
		span_end_html_index := current_html_index + span_length

		if match_start_index >= current_html_index && match_start_index+match_length <= span_end_html_index {
			utils.Log2("get overlapping html")
			overlapping_html := get_span_overlapping_html(span, match_start_index-current_html_index, match_length, match_html)
			result.WriteString(overlapping_html)
			match_start_index = int(^uint(0) >> 1) // Prevent further processing
		} else {
			// havent reached the match yet -> just add the outerhtml
			utils.Log2("get non-overlapping html")
			result.WriteString(span.OuterHtml)
		}
		current_html_index = span_end_html_index
	}
	return result.String()
}

func get_span_overlapping_html(span ChromaSpan, relative_start, match_length int, match_html string) string {
	var result bytes.Buffer
	relative_end := relative_start + match_length

	// HACK: trim previously whitespace for proper matching
	inner_html := span.InnerHtml
	inner_html = strings.TrimSpace(inner_html)
	span_length := len(inner_html)
	outer_html := strings.Replace(span.OuterHtml, span.InnerHtml, inner_html, 1)

	if relative_end > span_length {
		relative_end = span_length
	}

	if relative_start > 0 {
		result.WriteString(outer_html[:strings.Index(outer_html, inner_html)+relative_start])
	}

	result.WriteString(match_html)

	if relative_end < span_length {
		inner_html_start := strings.Index(outer_html, inner_html)
		after_match_start := inner_html_start + relative_end
		if after_match_start < len(outer_html) {
			result.WriteString(outer_html[after_match_start:])
		}
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
	utils.Log2(fmt.Sprintf("index of %s in combined innerHTML %s is", matched_text, current_text))
	startPos := strings.Index(current_text, matched_text)
	if startPos == -1 {
		return children, nil, nil
	}
	utils.Log2(fmt.Sprintf("index = %d", startPos))
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
