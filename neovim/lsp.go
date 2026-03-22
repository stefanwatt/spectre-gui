package neovim

type LspReferenceItem struct {
	StartCol     int    `msgpack:"col"`
	EndCol       int    `msgpack:"end_col"`
	StartRow     int    `msgpack:"lnum"`
	EndRow       int    `msgpack:"end_lnum"`
	AbsolutePath string `msgpack:"filename"`
	Text         string `msgpack:"text"`
}

type LspSymbolItem struct {
	StartCol     int    `msgpack:"col"`
	EndCol       int    `msgpack:"end_col"`
	StartRow     int    `msgpack:"lnum"`
	EndRow       int    `msgpack:"end_lnum"`
	AbsolutePath string `msgpack:"filename"`
	Text         string `msgpack:"text"`
	Kind         string `msgpack:"kind"`
}

func GetReferencesUnderCursor() []*LspReferenceItem {
	var references []*LspReferenceItem
	err := NvimInstance.ExecLua(`
        local items = {}
        local done = false
        
        -- Request references
        vim.lsp.buf.references(nil, {
            on_list = function(response)
                items = response.items or {}
                done = true
            end
        })
        
        -- Wait for the request to complete (with timeout)
        local timeout = 1000  -- milliseconds
        local start = vim.loop.now()
        while not done and (vim.loop.now() - start) < timeout do
            vim.wait(10)  -- Small sleep to avoid busy waiting
        end
        
        return items
    `, &references)

	assert(err == nil, "error getting lsp references")
	return references
}

func GetDocumentSymbols() []*LspSymbolItem {
	var symbols []*LspSymbolItem
	err := NvimInstance.ExecLua(`
        local items = {}
        local done = false
        
        -- Request references
        vim.lsp.buf.document_symbol({
            on_list = function(response)
                items = response.items or {}
                done = true
            end
        })
        
        -- Wait for the request to complete (with timeout)
        local timeout = 1000  -- milliseconds
        local start = vim.loop.now()
        while not done and (vim.loop.now() - start) < timeout do
            vim.wait(10)  -- Small sleep to avoid busy waiting
        end
        
        return items
    `, &symbols)

	assert(err == nil, "error getting lsp references")
	return symbols
}
