package _import

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportSet_String(t *testing.T) {
	tests := []struct {
		name     string
		imports  *ImportSet
		expected string
	}{
		{
			name:     "empty imports",
			imports:  NewImportSet(),
			expected: "",
		},
		{
			name: "single import",
			imports: func() *ImportSet {
				s := NewImportSet()
				s.Add(ImportDb)
				return s
			}(),
			expected: "import (\n  \"database/sql\"\n)\n",
		},
		{
			name: "multiple imports",
			imports: func() *ImportSet {
				s := NewImportSet()
				s.Add(ImportFMT)
				s.Add(ImportErrors)
				s.Add(ImportDb)
				return s
			}(),
			expected: "import (\n  \"fmt\"\n  \"github.com/pkg/errors\"\n  \"database/sql\"\n)\n",
		},
		{
			name: "imports with aliases",
			imports: func() *ImportSet {
				s := NewImportSet()
				s.Add(ImportSquirrel)
				s.Add(ImportLibPQ)
				return s
			}(),
			expected: "import (\n  sq \"github.com/Masterminds/squirrel\"\n  _ \"github.com/lib/pq\"\n)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.imports.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImportSet_Add(t *testing.T) {
	imports := NewImportSet()

	// Add single import
	imports.Add(ImportDb)
	assert.True(t, imports.imports[ImportDb])

	// Add multiple imports
	imports.Add(ImportFMT, ImportErrors)
	assert.True(t, imports.imports[ImportFMT])
	assert.True(t, imports.imports[ImportErrors])

	// Add duplicate import
	imports.Add(ImportDb)
	assert.True(t, imports.imports[ImportDb])
}

func TestImportSet_Enable(t *testing.T) {
	imports := NewImportSet()

	// Enable single import
	imports.Enable(ImportDb)
	assert.True(t, imports.imports[ImportDb])

	// Enable multiple imports
	imports.Enable(ImportFMT, ImportErrors)
	assert.True(t, imports.imports[ImportFMT])
	assert.True(t, imports.imports[ImportErrors])

	// Enable duplicate import
	imports.Enable(ImportDb)
	assert.True(t, imports.imports[ImportDb])
}

func TestImport_String(t *testing.T) {
	tests := []struct {
		name     string
		imp      Import
		expected string
	}{
		{
			name:     "import without alias",
			imp:      ImportDb,
			expected: "import  \"database/sql\"\n",
		},
		{
			name:     "import with alias",
			imp:      ImportSquirrel,
			expected: "import sq \"github.com/Masterminds/squirrel\"\n",
		},
		{
			name:     "import with underscore alias",
			imp:      ImportLibPQ,
			expected: "import _ \"github.com/lib/pq\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.imp.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImport_Alias(t *testing.T) {
	tests := []struct {
		name     string
		imp      Import
		expected string
	}{
		{
			name:     "import without alias",
			imp:      ImportDb,
			expected: "",
		},
		{
			name:     "import with alias",
			imp:      ImportSquirrel,
			expected: "sq",
		},
		{
			name:     "import with underscore alias",
			imp:      ImportLibPQ,
			expected: "_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.imp.alias()
			assert.Equal(t, tt.expected, result)
		})
	}
}
