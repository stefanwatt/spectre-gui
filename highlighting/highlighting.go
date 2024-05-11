package highlighting

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"spectre-gui/utils"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func Highlight(_code string, filename string, col int, matched_text string, replacement string) (string, string) {
	utils.Log(fmt.Sprintf("Highlighting code \n'%s' \nwith replacement: %s\non file%s", _code, replacement, filename))
	code, delimiter := add_delimiter(_code, col, matched_text)
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Get("plaintext")
		if lexer == nil {
			panic("Plaintext lexer not available")
		}
	}
	var bytes bytes.Buffer
	formatter := html.New(html.WithClasses(true), html.ClassPrefix("spectre-"))
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
	html := inject_match(highlighted, delimiter, matched_text, replacement)
	formatter.WriteCSS(&bytes, style)
	return html, bytes.String()
}

func inject_match(html string, delimiter string, matched_text string, replacement string) string {
	match_html := fmt.Sprintf("<span class=\"spectre-matched\">%s</span><span class=\"spectre-replacement\">%s</span>", matched_text, replacement)
	parts := strings.Split(html, delimiter)
	if len(parts) != 3 {
		panic("Failed to inject match into highlighted html")
	}
	return fmt.Sprintf("%s%s%s", parts[0], match_html, parts[2])
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
