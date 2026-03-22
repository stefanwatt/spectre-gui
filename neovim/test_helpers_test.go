package neovim

func resetHighlightState() {
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
	effectiveHlIdsMu.Unlock()

	currentFgColorId = 1
	currentBgColorId = 1
}
