package rendering

import (
	"strings"

	"github.com/akiyosi/goneovim/util"
)

type Cmdline struct {
	Content string `json:"content"`
	Pos     int    `json:"pos"`
	Firstc  string `json:"firstc"`
	Prompt  string `json:"prompt"`
	Indent  int    `json:"indent"`
}

func RenderCmdline(arg []interface{}) Cmdline {
	content := ""
	contentChunks := arg[0].([]interface{})
	for _, e := range contentChunks {
		a := e.([]interface{})

		if len(a) < 2 {
			// content += a[0].(string)
			content += strings.Replace(a[0].(string), "\t", " ", -1)
		} else {
			if len(contentChunks) == 1 {
				// content += a[1].(string)
				content += strings.Replace(a[1].(string), "\t", " ", -1)
			} else {
				content += sanitize(a[1].(string))
			}
		}
	}
	// content := arg[0].([]interface{})[0].([]interface{})[1].(string)

	return Cmdline{
		Content: content,
		Pos:     util.ReflectToInt(arg[1]),
		Firstc:  arg[2].(string),
		Prompt:  arg[3].(string),
		Indent:  util.ReflectToInt(arg[4]),
	}
}
