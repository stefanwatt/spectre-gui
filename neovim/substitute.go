package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"regexp"
	"sync"
)

var (
	lastSearchPos  int = 0
	lastReplacePos int = 0
	lastSearchLen  int = -69
	positionMutex  sync.Mutex
)

func assert(assertion bool, message string) {
	if !assertion {
		panic(message)
	}
}

// TODO: i reintroduced the bug when cursor is at the end of the replace field for example
func HandleSubstituteJump(optionalData ...interface{}) {
	utils.Log("HandleSubstituteJump")
	var cmdlineContentRes interface{}
	var cmdlinePosRes interface{}

	NvimInstance.ExecLua("return vim.fn.getcmdline()", &cmdlineContentRes)
	NvimInstance.ExecLua("return vim.fn.getcmdpos()", &cmdlinePosRes)

	cmdlineContent, ok := cmdlineContentRes.(string)
	if !ok {
		utils.Log("HandleSubstituteJump couldnt reflect cmdline content")
		return
	}
	cmdlinePos := utils.ReflectToInt(cmdlinePosRes)

	match := regexp.MustCompile(`^(.*?)s\/(.*?)\/(.*?)(?:\/([gicI]*))?$`).FindStringSubmatch(cmdlineContent)
	if match == nil {
		return
	}

	rangeStr := match[1]
	search := match[2]
	replace := match[3]

	searchStart := len(rangeStr) + 2 // After "s/"
	searchEnd := searchStart + len(search)
	replaceStart := searchEnd + 1 // After the second "/"
	replaceEnd := replaceStart + len(replace)
	currentPos := cmdlinePos
	if lastSearchLen == -69 {
		lastSearchLen = searchEnd - searchStart
	}

	searchLen := searchEnd - searchStart
	assert(searchStart <= searchEnd && searchEnd <= replaceStart && replaceStart <= replaceEnd, "invalid substitute positions")
	if lastSearchPos == 0 {
		lastSearchPos = searchStart + 1
	}
	if lastReplacePos == 0 {
		lastReplacePos = replaceStart + 1
	}
	utils.Log(fmt.Sprintf("HandleSubstituteJump searchStart=%d searchEnd=%d replaceStart=%d replaceEnd=%d currentPos=%d", searchStart, searchEnd, replaceStart, replaceEnd, currentPos))

	positionMutex.Lock()
	defer positionMutex.Unlock()

	isInSearchField := currentPos > searchStart && currentPos <= searchEnd
	isInReplaceField := currentPos > replaceStart && currentPos <= replaceEnd

	if isInSearchField || isInReplaceField {
		assert(isInSearchField != isInReplaceField, "cant be in search and replace field at the same time")
	}

	var newPos int
	if isInSearchField {
		lastSearchPos = currentPos
		newPos = lastReplacePos
		if lastSearchLen != searchLen {
			newPos += (searchLen - lastSearchLen)
		}
	} else if isInReplaceField {
		lastReplacePos = currentPos
		newPos = lastSearchPos
	}
	var res interface{}
	lastSearchLen = searchLen
	NvimInstance.ExecLua(fmt.Sprintf("return vim.fn.setcmdline('%s',%d)", cmdlineContent, newPos), &res)
	utils.Log("HandleSubstituteJump response: ", res)
}
