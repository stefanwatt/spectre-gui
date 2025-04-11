package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	FG_COLOR_CLASSES_FILE = "/home/stefan/.config/nvim-gui/fg-color-to-class.txt"
	BG_COLOR_CLASSES_FILE = "/home/stefan/.config/nvim-gui/bg-color-to-class.txt"
	ID_CLASSES_FILE       = "/home/stefan/.config/nvim-gui/id-to-classes.txt"
	ENTRY_SEPARATOR       = "|"
	CLASS_SEPARATOR       = ","
	HEX_COLOR_REGEX       = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	dbReady               = make(chan struct{})
	dbReadyClosed         = false
	hlAttrQueue           = make(chan func())
	hlAttrQueueClosed     = false
	wg                    sync.WaitGroup
	isDBLoaded            atomic.Bool
	currentFgColorId      = 1
	currentBgColorId      = 1
	fgColorClassesMu      sync.Mutex
	bgColorClassesMu      sync.Mutex
	idClassesMu           sync.Mutex
	fgColorFileOpMu       sync.Mutex
	bgColorFileOpMu       sync.Mutex
	idFileOpMu            sync.Mutex
)

func createFileIfNotExist(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644)
	if err != nil {
		if os.IsExist(err) {
			fmt.Printf("File '%s' already exists\n", filename)
			return nil
		}
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	fmt.Printf("File '%s' created successfully\n", filename)
	return nil
}

func loadDatabase() {
	utils.Log("loading db...")
	var err error
	fgColorClasses, err = getForegroundColorClasses()
	if err != nil {
		panic(err.Error())
	}
	bgColorClasses, err = getBackgroundColorClasses()
	if err != nil {
		panic(err.Error())
	}
	idClasses, err = getIdClasses()
	if err != nil {
		panic(err.Error())
	}
	if !dbReadyClosed {
		close(dbReady)
		dbReadyClosed = true
	}

	isDBLoaded.Store(true)
}
func processHlAttrQueue() {
	<-dbReady

	utils.Log("db loaded... starting to process hl attr define queue")
	for fn := range hlAttrQueue {
		fn()
		wg.Done()
	}
}

func waitForHlAttrDefine() {
	if hlAttrQueueClosed {
		return
	}
	close(hlAttrQueue)
	hlAttrQueueClosed = true
	wg.Wait()
}

func isValidHexColor(color string) bool {
	return HEX_COLOR_REGEX.MatchString(color)
}

func addBackgroundColorClass(color string) error {
	if strings.TrimSpace(color) == "" {
		return nil
	}
	assert(isValidHexColor(color), fmt.Sprintf("invalid hex color=%slen(color)=%d", color, len(color)))

	bgColorFileOpMu.Lock()
	defer bgColorFileOpMu.Unlock()

	bgColorClassesMu.Lock()
	_, exists := bgColorClasses[color]
	bgColorClassesMu.Unlock()

	if exists {
		return nil
	}

	class := fmt.Sprintf("bg-%d", currentBgColorId)
	utils.Log(fmt.Sprintf("adding class %s for color key %s in db file", class, color))
	currentBgColorId++

	value := fmt.Sprintf("%s%s%s\n", color, ENTRY_SEPARATOR, class)

	err := appendStringToFile(BG_COLOR_CLASSES_FILE, value)
	if err == nil {
		bgColorClassesMu.Lock()
		bgColorClasses[color] = class
		bgColorClassesMu.Unlock()
	}
	return err
}

func addForegroundColorClass(color string) error {
	if strings.TrimSpace(color) == "" {
		return nil
	}
	assert(isValidHexColor(color), fmt.Sprintf("invalid hex color=%slen(color)=%d", color, len(color)))

	fgColorFileOpMu.Lock()
	defer fgColorFileOpMu.Unlock()

	fgColorClassesMu.Lock()
	_, exists := fgColorClasses[color]
	fgColorClassesMu.Unlock()

	if exists {
		return nil
	}

	class := fmt.Sprintf("fg-%d", currentFgColorId)
	utils.Log(fmt.Sprintf("adding class %s for color key %s in db file", class, color))
	currentFgColorId++

	value := fmt.Sprintf("%s%s%s\n", color, ENTRY_SEPARATOR, class)

	err := appendStringToFile(FG_COLOR_CLASSES_FILE, value)
	if err == nil {
		fgColorClassesMu.Lock()
		fgColorClasses[color] = class
		fgColorClassesMu.Unlock()
	}
	return err
}

func getForegroundColorClasses() (map[string]string, error) {
	lines, err := readLinesFromFile(FG_COLOR_CLASSES_FILE)
	if err != nil {
		return nil, err
	}
	currentFgColorId = len(lines) + 1
	return getColorClasses(lines)
}

func getBackgroundColorClasses() (map[string]string, error) {
	lines, err := readLinesFromFile(BG_COLOR_CLASSES_FILE)
	if err != nil {
		return nil, err
	}

	currentBgColorId = len(lines) + 1
	return getColorClasses(lines)
}

func getColorClasses(lines []string) (map[string]string, error) {
	colorClasses := make(map[string]string)
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return colorClasses, nil
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ENTRY_SEPARATOR)
		assert(len(parts) == 2, fmt.Sprintf("malformed color classes entry len(parts)=%d\nline:%s", len(parts), line))
		key := parts[0]
		class := parts[1]
		colorClasses[key] = class
	}
	utils.Log(fmt.Sprintf("Loaded %d color classes. Next color ID: %d", len(colorClasses), currentFgColorId))
	return colorClasses, nil
}

func addIdClasses(id int, classes []string) error {
	idFileOpMu.Lock()
	defer idFileOpMu.Unlock()
	idClassesMu.Lock()
	defer idClassesMu.Unlock()
	_, exists := idClasses[id]
	if exists {
		return nil
	}
	value := fmt.Sprintf("%d|%s\n", id, strings.Join(classes, ","))
	err := appendStringToFile(ID_CLASSES_FILE, value)
	if err != nil {
		return err
	}
	idClasses[id] = classes
	return nil
}

func getIdClasses() (map[int][]string, error) {
	lines, err := readLinesFromFile(ID_CLASSES_FILE)
	if err != nil {
		return nil, err
	}
	idClasses := make(map[int][]string)

	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return idClasses, nil
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ENTRY_SEPARATOR)
		assert(len(parts) == 2, "malformed id classes entry")
		classes := strings.Split(parts[1], CLASS_SEPARATOR)
		id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		assert(err == nil, "malformed id: not an int: "+parts[0])
		idClasses[id] = classes
		effectiveHlIdsMu.Lock()
		effectiveHlIds[mapClassesString(classes)] = id
		effectiveHlIdsMu.Unlock()
	}
	return idClasses, nil
}

func readLinesFromFile(filepath string) ([]string, error) {
	err := createFileIfNotExist(filepath)
	if err != nil {
		return nil, err
	}
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), "\n"), nil
}

func appendStringToFile(filepath string, value string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(value)
	return err
}
