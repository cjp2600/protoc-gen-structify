package plugin

import (
	"fmt"
	"strings"
)

// ImportSet is a set of imports.
type ImportSet map[Import]bool

// String returns a string representation of the ImportSet.
func (i ImportSet) String() string {
	if len(i) > 0 {
		builder := &strings.Builder{}

		builder.WriteString("import (\n")
		for k := range i {
			builder.WriteString(strings.ReplaceAll(k.String(), "import", ""))
		}
		builder.WriteString(")\n")

		return builder.String()
	}

	var out string
	for k := range i {
		out += k.String()
	}

	return out
}

// Add adds imports to the ImportSet.
func (i ImportSet) Add(imports ...Import) {
	for _, v := range imports {
		i[v] = false
	}
}

// Enable enables imports in the ImportSet.
func (i ImportSet) Enable(imports ...Import) {
	for _, v := range imports {
		i[v] = true
	}
}

// Import is a type for how to generate import paths.
type Import string

func (i Import) String() string {
	return fmt.Sprintf("import %s \"%s\"\n", i.alias(), string(i))
}

func (i Import) alias() string {
	switch i {
	case ImportLibPQ:
		return "_"
	case ImportSquirrel:
		return "sq"
	default:
		return ""
	}
}

var (
	ImportDb       = Import("database/sql")
	ImportLibPQ    = Import("github.com/lib/pq")
	ImportStrings  = Import("strings")
	ImportSquirrel = Import("github.com/Masterminds/squirrel")
	ImportFMT      = Import("fmt")
)
