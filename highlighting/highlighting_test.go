package highlighting

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

var (
	insert_style = lipgloss.NewStyle().Foreground(lipgloss.Color("#303446")).Background(lipgloss.Color("#a6d189"))
	delete_style = lipgloss.NewStyle().Foreground(lipgloss.Color("#303446")).Background(lipgloss.Color("#e78284"))
	equal_style  = lipgloss.NewStyle().Foreground(lipgloss.Color("#c6d0f5")).Background(lipgloss.Color("#303446"))
)

func TestHighlight(t *testing.T) {
	t.Run("Highlight across tokens with regex", test_highlight_regex)
	t.Run("Highlight code", test_highlight_code)
	t.Run("Test find body node", test_find_body_node)
	t.Run("Test inject match", test_inject_match)
}

func test_find_body_node(t *testing.T) {
	highlightedHTML := `<html><head></head><body><code class="spectre-chroma">    <span class="spectre-nx">prepare_node</span> <span class="spectre-p">=</span> <span class="spectre-nf">function</span><span class="spectre-p">(</span><span class="spectre-nx">node</span><span class="spectre-p">)</span></code></body></html>`
	doc, err := html.Parse(strings.NewReader(highlightedHTML))
	assert.Nil(t, err)

	actual := find_body_node(doc)
	assert.NotNil(t, actual)

	expected := `<body><code class="spectre-chroma">    <span class="spectre-nx">prepare_node</span> <span class="spectre-p">=</span> <span class="spectre-nf">function</span><span class="spectre-p">(</span><span class="spectre-nx">node</span><span class="spectre-p">)</span></code></body>`
	var buffer strings.Builder
	err = html.Render(&buffer, actual)
	assert.Nil(t, err)
	assert.Equal(t, expected, buffer.String())
}

func test_inject_match(t *testing.T) {
	highlighted_html := `<code class="spectre-chroma">    <span class="spectre-nx">prepare_node</span> <span class="spectre-p">=</span> <span class="spectre-nf">function</span><span class="spectre-p">(</span><span class="spectre-nx">node</span><span class="spectre-p">)</span></code>`
	matched_text := `function(node)`
	match_html := `fn(node)`

	actual := inject_match(highlighted_html, matched_text, match_html)

	expected := `<code class="spectre-chroma">    <span class="spectre-nx">prepare_node</span> <span class="spectre-p">=</span> <span class="spectre-nf"><span class="spectre-matched">function(node)</span><span class="spectre-replacement">fn(node)</span></span></code>`
	assert.Equal(t, expected, actual)
}

func test_highlight_regex(t *testing.T) {
	var (
		code         = `    prepare_node = function(node)`
		filename     = `foo.go`
		matched_text = `function(node)`
		replacement  = `fn(node))`
	)

	actual, _ := Highlight(code, filename, matched_text, replacement)

	expected := `<code class="spectre-chroma">    <span class="spectre-nx">prepare_node</span> <span class="spectre-p">=</span> <span class="spectre-nf"><span class="spectre-matched">function(node)</span><span class="spectre-replacement">fn(node))</span></span></code>`
	if expected != actual {
		assert.Equal(t, expected, actual, "HTML output did not match expected.")
	}
}

func test_highlight_code(t *testing.T) {
	var (
		code     = `    prepare_node = function(node)`
		filename = `foo.go`
	)
	actual := highlight_code(code, filename)
	expected := `<code class="spectre-chroma">    <span class="spectre-nx">prepare_node</span> <span class="spectre-p">=</span> <span class="spectre-nf">function</span><span class="spectre-p">(</span><span class="spectre-nx">node</span><span class="spectre-p">)</span></code>`
	if expected != actual {

		expected = strings.ReplaceAll(expected, ">", ">\n")
		actual = strings.ReplaceAll(actual, ">", ">\n")
		assert.Equal(t, expected, actual, "HTML output did not match expected.")
	}
}
