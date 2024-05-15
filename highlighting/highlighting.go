package highlighting

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"spectre-gui/utils"

	"github.com/alecthomas/chroma/v2"
	html_formatter "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func Highlight(code, filename, matched_text, replacement string) (string, string) {
	utils.Log2(fmt.Sprintf("Highlighting code \n'%s' \nwith replacement: %s\non file %s", code, replacement, filename))
	highlighted := highlight_code(code, filename)
	utils.Log2(fmt.Sprintf("highlighted code: %s", highlighted))
	html := inject_match(highlighted, matched_text, replacement)
	return html, highlighted
}

func highlight_code(code string, filename string) string {
	var buf bytes.Buffer

	lexer := match_lexer(filename)
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		log.Fatal(err)
	}
	formatter := html_formatter.New(html_formatter.WithClasses(true), html_formatter.ClassPrefix("spectre-"), html_formatter.InlineCode(true))
	style := styles.Get("catppuccin-frappe")
	if style == nil {
		style = styles.Fallback
	}
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		log.Fatal(err)
	}
	highlighted := buf.String()
	return highlighted
}

var lexerCache = make(map[string]chroma.Lexer)

func match_lexer(filename string) chroma.Lexer {
	ext := strings.TrimLeft(filepath.Ext(filename), ".")
	if cachedLexer, exists := lexerCache[ext]; exists {
		return cachedLexer
	} else {
		lexer := lexers.Get(ext)
		if lexer == nil {
			lexer = lexers.Fallback
		}
		lexerCache[ext] = lexer
		return lexer
	}
}

func inject_match(highlighted_html, matched_text, replacement string) string {
	if matched_text == "" {
		log.Printf("matched text empty -> cannot inject match html")
		return highlighted_html
	}

	fmt.Println(fmt.Sprintf("[inject match] mathched_text: %s\nreplacement: %s", matched_text, replacement))
	match_html := fmt.Sprintf(
		`<span class="spectre-matched">%s</span><span class="spectre-replacement">%s</span>`,
		matched_text,
		replacement,
	)

	characters := strings.Split(matched_text, "")
	characters = utils.MapArray(characters, regexp.QuoteMeta)
	// Allow for arbitrary HTML tags between characters
	matched_text_pattern := strings.Join(characters, "(?:<[^>]*?>)*?")
	matched_text_re := regexp.MustCompile(matched_text_pattern)

	if matched_text_re.MatchString(highlighted_html) && !strings.Contains(highlighted_html, match_html) {
		fmt.Println(fmt.Sprintf("[inject match] match_html: %s", match_html))
		fmt.Println(fmt.Sprintf("[inject match] highlighted_html: %s", highlighted_html))
		injected_html := matched_text_re.ReplaceAllStringFunc(highlighted_html, func(match string) string {
			// Ensure the match starts and ends at appropriate positions
			fmt.Println(fmt.Sprintf("[inject match] match: %s, matched_text: %s", match, matched_text))
			return match_html
		})
		fmt.Println(fmt.Sprintf("[inject match] injected_html: %s", injected_html))
		return injected_html
	}
	return highlighted_html
}

func find_body_node(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := find_body_node(c); result != nil {
			return result
		}
	}
	return nil
}
