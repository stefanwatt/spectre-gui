package neovim

import (
	"strings"
)

type QuickfixEntry struct {
	Filepath string `json:"filepath" msgpack:"filename"`
	Row      int    `json:"row" msgpack:"lnum"`
	Col      int    `json:"col" msgpack:"col"`
	Text     string `json:"text" msgpack:"text"`
}

func SetQuickfixList(entries []*QuickfixEntry) {
	// log.Debug(fmt.Sprintf("AppendToQuickfixList filepath=%s row=%d col=%d text=%s", entry.Filepath, entry.Row, entry.Col, entry.Text))
	err := NvimClient.Command("copen")
	assert(err == nil, "error opening qfl")
	var res interface{}
	for _, entry := range entries {
		entry.Text = strings.TrimSpace(strings.ReplaceAll(entry.Text, "\x00", ""))
	}
	err = NvimClient.ExecLua(`
        local entries = ...
        return vim.fn.setqflist(entries, 'r')
    `, &res, entries)

	assert(err == nil, "error adding entry to qfl")
}
