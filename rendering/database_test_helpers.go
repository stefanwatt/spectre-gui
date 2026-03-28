package rendering

func resetHighlightState() {
	currentFgColorId = 1
	currentBgColorId = 1
	idClasses = make(map[int][]string)
	fgColorClasses = make(map[string]string)
	bgColorClasses = make(map[string]string)
	effectiveHlIds = make(map[string]int)
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
