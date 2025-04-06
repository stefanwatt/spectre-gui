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
	COLOR_CLASSES_FILE = "/home/stefan/.config/nvim-gui/color-to-class.txt"
	ID_CLASSES_FILE    = "/home/stefan/.config/nvim-gui/id-to-classes.txt"
	ENTRY_SEPARATOR    = "|"
	CLASS_SEPARATOR    = ","
	HEX_COLOR_REGEX    = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	dbReady            = make(chan struct{})
	hlAttrQueue        = make(chan func())
	wg                 sync.WaitGroup
	isDBLoaded         atomic.Bool
	currentColorId     = 1
	colorClassesMu     sync.Mutex
	idClassesMu        sync.Mutex
	colorFileOpMu      sync.Mutex
	idFileOpMu         sync.Mutex
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
	colorClasses, err = getColorClasses()
	if err != nil {
		panic(err.Error())
	}
	idClasses, err = getIdClasses()
	if err != nil {
		panic(err.Error())
	}
	close(dbReady)
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
	close(hlAttrQueue)
	wg.Wait()
}

func isValidHexColor(color string) bool {
	return HEX_COLOR_REGEX.MatchString(color)
}

func addColorClass(color string, colorTypePrefix string) error {
	if strings.TrimSpace(color) == "" {
		return nil
	}
	assert(isValidHexColor(color), fmt.Sprintf("invalid hex color=%slen(color)=%d", color, len(color)))

	colorFileOpMu.Lock()
	defer colorFileOpMu.Unlock()

	colorClassesMu.Lock()
	_, exists := colorClasses[color]
	colorClassesMu.Unlock()
	class := fmt.Sprintf("%s-%d", colorTypePrefix, currentColorId)
	if exists {
		return nil
	}
	utils.Log(fmt.Sprintf("adding class for color %s in db file", color))
	currentColorId++

	value := fmt.Sprintf("%s%s%s\n", color, ENTRY_SEPARATOR, class)

	err := appendStringToFile(COLOR_CLASSES_FILE, value)
	if err == nil {
		colorClassesMu.Lock()
		colorClasses[color] = class
		colorClassesMu.Unlock()
	}
	return err
}

func getColorClasses() (map[string]string, error) {
	lines, err := readLinesFromFile(COLOR_CLASSES_FILE)
	if err != nil {
		return nil, err
	}
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
		colorClasses[parts[0]] = parts[1]
		currentColorId++
	}
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
