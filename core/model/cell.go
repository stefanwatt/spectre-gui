package model

import (
	"sort"
	"strings"
)

type Cell struct {
	Char      string
	Highlight int
	Dirty     bool
	Classes   map[string]bool
}

func (c *Cell) ClassesToString() string {
	var builder strings.Builder
	classes := make([]string, 0, len(c.Classes))
	for class := range c.Classes {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	for _, class := range classes {
		builder.WriteString(class)
		builder.WriteString(" ")
	}
	return strings.TrimRight(builder.String(), " ")
}

func (c *Cell) Equals(other *Cell) bool {
	if c == nil || other == nil {
		return c == other
	}
	return c.Char == other.Char && c.ClassesToString() == other.ClassesToString() && c.Highlight == other.Highlight
}
