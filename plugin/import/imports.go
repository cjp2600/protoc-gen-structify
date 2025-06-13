package _import

import (
	"fmt"
	"strings"
)

// ImportSet is a set of imports.
type ImportSet struct {
	imports map[Import]bool
	order   []Import
}

// NewImportSet creates a new ImportSet.
func NewImportSet() *ImportSet {
	return &ImportSet{
		imports: make(map[Import]bool),
		order:   make([]Import, 0),
	}
}

// String returns a string representation of the ImportSet.
func (i *ImportSet) String() string {
	if len(i.imports) > 0 {
		builder := &strings.Builder{}

		builder.WriteString("import (\n")
		for _, imp := range i.order {
			builder.WriteString("  ")
			importStr := strings.TrimSpace(strings.ReplaceAll(imp.String(), "import", ""))
			builder.WriteString(importStr)
			builder.WriteString("\n")
		}
		builder.WriteString(")\n")

		return builder.String()
	}

	return ""
}

// Add adds imports to the ImportSet.
func (i *ImportSet) Add(imports ...Import) {
	for _, v := range imports {
		if !i.imports[v] {
			i.imports[v] = true
			i.order = append(i.order, v)
		}
	}
}

// Enable enables imports in the ImportSet.
func (i *ImportSet) Enable(imports ...Import) {
	for _, v := range imports {
		if !i.imports[v] {
			i.imports[v] = true
			i.order = append(i.order, v)
		}
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
	ImportClickhouse        = Import{"github.com/ClickHouse/clickhouse-go/v2", ""}
	ImportClickhouseDriver  = Import{"github.com/ClickHouse/clickhouse-go/v2/lib/driver", ""}
)

// GetImports returns a slice of all imports in the ImportSet.
func (i *ImportSet) GetImports() []Import {
	return i.order
}
