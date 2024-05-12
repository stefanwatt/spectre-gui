package lua

import (
	"fmt"
	"os"

	golua "github.com/yuin/gopher-lua"
)

type Config struct {
	CaseSensitive  bool
	Regex          bool
	MatchWholeWord bool
	PreserveCase   bool
}

// TODO: build path
var CONFIG_PATH = "/home/stefan/.config/spectre-gui/init.lua"

func LoadConfig() Config {
	L := golua.NewState()
	defer L.Close()

	// check if fil exists
	_, err := os.Stat(CONFIG_PATH)
	if err != nil {
		// TODO: copy default config
		panic("init.lua not found")
	}
	if err = L.DoFile(CONFIG_PATH); err != nil {
		panic(err)
	}

	var config Config
	lua_config := L.Get(-1)
	if tbl, ok := lua_config.(*golua.LTable); ok {
		case_sensitive := tbl.RawGetString("case_sensitive")
		config.CaseSensitive = golua.LVAsBool(case_sensitive)
		regex := tbl.RawGetString("regex")
		config.Regex = golua.LVAsBool(regex)
		match_whole_word := tbl.RawGetString("match_whole_word")
		config.MatchWholeWord = golua.LVAsBool(match_whole_word)
		preserveCase := tbl.RawGetString("preserve_case")
		config.PreserveCase = golua.LVAsBool(preserveCase)
	} else {
		panic(fmt.Sprintf("config.lua must return a table, got %T", lua_config))
	}
	return config
}
