package rendering

// ResetHighlightStateForTest resets all highlight state for testing.
// Exported so other packages' tests (e.g. neovim) can reset rendering state.
func ResetHighlightStateForTest() {
	fgColorClassesMu.Lock()
	fgColorClasses = make(map[string]string)
	fgColorClassesMu.Unlock()

	bgColorClassesMu.Lock()
	bgColorClasses = make(map[string]string)
	bgColorClassesMu.Unlock()

	idClassesMu.Lock()
	idClasses = make(map[int][]string)
	idClassesMu.Unlock()

	effectiveHlIdsMu.Lock()
	effectiveHlIds = make(map[string]int)
	hlIdToEffective = make(map[int]int)
	effectiveHlIdsMu.Unlock()

	currentFgColorId = 1
	currentBgColorId = 1
}

func resetHighlightState() {
	ResetHighlightStateForTest()
}

func addForegroundColorClass(color string) {
	AddForegroundColorClass(color)
}

func addBackgroundColorClass(color string) {
	AddBackgroundColorClass(color)
}

func addIdClasses(id int, classes []string) {
	AddIdClasses(id, classes)
}
