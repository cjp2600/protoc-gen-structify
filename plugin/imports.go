package plugin

// Import is a type for how to generate import paths.
type Import string

func (i Import) String() string {
	return string(i)
}

var (
	ImportSqlx = Import("github.com/jmoiron/sqlx")
	ImportDb   = Import("database/sql")
)
