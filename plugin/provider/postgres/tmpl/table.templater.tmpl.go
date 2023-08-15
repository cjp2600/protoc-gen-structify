package tmpl

const TableTemplate = `
{{ template "storage" . }}
{{ template "structure" . }}
`

const TableStorageTemplate = `
// {{ storageName | lowerCamelCase }} is a struct for the "{{ tableName }}" table.
type {{ storageName | lowerCamelCase }} struct {
	db *sql.DB // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

type {{ storageName }} interface {
	// CreateTable creates the table.
	CreateTable() error
	// DropTable drops the table.
	DropTable() error
	// TruncateTable truncates the table.
	TruncateTable() error
	// UpgradeTable upgrades the table.
	UpgradeTable() error
}

// New{{ storageName }} returns a new {{ storageName | lowerCamelCase }}.
func New{{ storageName }}(db *sql.DB) {{ storageName }} {
	return &{{ storageName | lowerCamelCase }}{
		db: db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// createTable creates the table.
func (t *{{ storageName | lowerCamelCase }}) CreateTable() error {
	sqlQuery := ` + "`" + `
		CREATE TABLE IF NOT EXISTS {{ tableName }} (
		{{- range $index, $field := fields }}
		{{- if not ($field | isRelation) }}
		{{ $field | sourceName }} {{if ($field | isAutoIncrement ) }}{{ $field | postgresType }} SERIAL{{else}}{{ $field | postgresType }}{{end}}{{if $field | isPrimaryKey }} PRIMARY KEY{{end}}{{if ($field | isUnique) }} UNIQUE{{end}}{{if ($field | isNotNull ) }} NOT NULL{{end}}{{if ($field | getDefaultValue) }} DEFAULT {{$field | getDefaultValue}}{{end}}{{if not ( $field | isLastField )}},{{end}}
		{{- end}}
		{{- end}});
		{{if (comment) }}COMMENT ON TABLE {{ tableName }} IS '{{ comment }}';{{end}}
	` + "`" + `

	_, err := t.db.Exec(sqlQuery)
	return err
}

// DropTable drops the table.
func (t *{{ storageName | lowerCamelCase }}) DropTable() error {
	sqlQuery := ` + "`" + `
		DROP TABLE IF EXISTS {{ tableName }};
	` + "`" + `

	_, err := t.db.Exec(sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *{{ storageName | lowerCamelCase }}) TruncateTable() error {
	sqlQuery := ` + "`" + `
		TRUNCATE TABLE {{ tableName }};
	` + "`" + `

	_, err := t.db.Exec(sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
func (t *{{ storageName | lowerCamelCase }}) UpgradeTable() error {
	return nil
}
`

const StructureTemplate = `
// {{ structureName }} is a struct for the "{{ tableName }}" table.
type {{ structureName }} struct {
{{ range $field := fields }}
	{{ $field | fieldName }} {{ $field | fieldType }}{{if not ($field | isRelation) }}` + " `db:\"{{ $field | sourceName }}\"`" + `{{end}}{{end}}
}

// ScanRow scans a row into a {{ structureName }}.
func (t *{{ structureName }}) ScanRow(r *sql.Row) error {
	return r.Scan({{ range $field := fields }} {{if not ($field | isRelation) }} &t.{{ $field | fieldName }}, {{ end }}{{ end }})
}

// TableName returns the table name.
func (t *{{ structureName }}) TableName() string {
	return "{{ tableName }}"
}

// Columns returns the columns for the table.
func (t *{{ structureName }}) Columns() []string {
	return []string{
		{{ range $field := fields }}{{if not ($field | isRelation) }}"{{ $field | sourceName }}",{{ end }}{{ end }}
	}
}
`
