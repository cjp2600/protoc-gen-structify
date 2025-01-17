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
		i[v] = true
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

// String returns a string representation of the Import.
func (i Import) String() string {
	return fmt.Sprintf("import %s \"%s\"\n", i.alias(), i.path)
}

// alias returns the alias for the import.
func (i Import) alias() string {
	return i.sub
}

var (
	ImportDb                = Import{"database/sql", ""}
	ImportLibPQ             = Import{"github.com/lib/pq", "_"}
	ImportLibPQWOAlias      = Import{"github.com/lib/pq", ""}
	ImportLibSqlite3        = Import{"github.com/mattn/go-sqlite3", "_"}
	ImportLibSqlite3WOAlias = Import{"github.com/mattn/go-sqlite3", ""}
	ImportStrings           = Import{"strings", ""}
	ImportMath              = Import{"math", ""}
	ImportSquirrel          = Import{"github.com/Masterminds/squirrel", "sq"}
	ImportNull              = Import{"gopkg.in/guregu/null.v4", ""}
	ImportFMT               = Import{"fmt", ""}
	ImportErrors            = Import{"github.com/pkg/errors", ""}
	ImportContext           = Import{"context", ""}
	ImportStrconv           = Import{"strconv", ""}
	ImportSync              = Import{"sync", ""}
	ImportTime              = Import{"time", ""}
	ImportJson              = Import{"encoding/json", ""}
	ImportSQLDriver         = Import{"database/sql/driver", ""}
	ImportGoogleUUID        = Import{"github.com/google/uuid", ""}
)
