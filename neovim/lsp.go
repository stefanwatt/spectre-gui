package neovim

type LspReferenceItem struct {
	StartCol     int    `msgpack:"col"`
	EndCol       int    `msgpack:"end_col"`
	StartRow     int    `msgpack:"lnum"`
	EndRow       int    `msgpack:"end_lnum"`
	AbsolutePath string `msgpack:"filename"`
	Text         string `msgpack:"text"`
}

func GetReferencesUnderCursor() []*LspReferenceItem {
	var references []*LspReferenceItem
	err := NvimInstance.ExecLua(`
		local items = {}
		vim.lsp.buf.references(nil,{on_list=function(response) items=response.items end})
		return items
    `, &references)
	assert(err == nil, "error getting lsp references")
	return references
}
