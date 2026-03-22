package neovim

import "nvim-gui/utils"

type HelpTag struct {
	Tag      string `msgpack:"tag"`
	Filename string `msgpack:"filename"`
	Filepath string `msgpack:"filepath"`
	Cmd      string `msgpack:"cmd"`
	Lang     string `msgpack:"lang"`
}

func GetHelpTags() []*HelpTag {
	var helpTags []*HelpTag
	err := NvimInstance.ExecLua("return require('config.help-tags').get_help_tags()", &helpTags)
	assert(err == nil, "error getting help tags")
	utils.Log("GetHelpTags", helpTags)
	return helpTags
}
