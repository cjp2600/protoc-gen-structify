package _import

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
type Import struct {
	path string
	sub  string
}

func (i Import) String() string {
	return fmt.Sprintf("import %s \"%s\"\n", i.alias(), i.path)
}

func (i Import) alias() string {
	return i.sub
}

var (
	ImportDb           = Import{"database/sql", ""}
	ImportLibPQ        = Import{"github.com/lib/pq", "_"}
	ImportLibPQWOAlias = Import{"github.com/lib/pq", ""}
	ImportStrings      = Import{"strings", ""}
	ImportSquirrel     = Import{"github.com/Masterminds/squirrel", "sq"}
	ImportFMT          = Import{"fmt", ""}
	ImportErrors       = Import{"errors", ""}
	ImportContext      = Import{"context", ""}
	ImportStrconv      = Import{"strconv", ""}
	ImportSync         = Import{"sync", ""}
	ImportTime         = Import{"time", ""}
	ImportJson         = Import{"encoding/json", ""}
	ImportSQLDriver    = Import{"database/sql/driver", ""}
)
