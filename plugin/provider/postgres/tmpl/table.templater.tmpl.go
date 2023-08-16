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
	// Create creates a new {{ structureName }}.
	{{ if (hasID) }}Create(ctx context.Context, model *{{structureName}}) (*{{IDType}}, error)
	{{ else }} Create(ctx context.Context, model *{{structureName}}) error
	{{ end }}
}

// New{{ storageName }} returns a new {{ storageName | lowerCamelCase }}.
func New{{ storageName }}(db *sql.DB) {{ storageName }} {
	return &{{ storageName | lowerCamelCase }}{
		db: db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *{{ storageName | lowerCamelCase }}) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *{{ storageName | lowerCamelCase }}) CreateTable() error {
	sqlQuery := ` + "`" + `
		{{- range $index, $field := fields }}
		{{- if ($field | isDefaultUUID ) }}
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		{{- end}}
		{{- end}}
		-- Table: {{ tableName }}
		CREATE TABLE IF NOT EXISTS {{ tableName }} (
		{{- range $index, $field := fields }}
		{{- if not ($field | isRelation) }}
		{{ $field | sourceName }} {{if ($field | isAutoIncrement) }} SERIAL{{else}}{{ $field | postgresType }}{{end}}{{if $field | isPrimaryKey }} PRIMARY KEY{{end}}{{if ($field | isUnique) }} UNIQUE{{end}}{{ if and (isNotNull $field) (not (isAutoIncrement $field)) }} NOT NULL{{ end }}{{if ($field | getDefaultValue) }} DEFAULT {{$field | getDefaultValue}}{{end}}{{if not ( $field | isLastField )}},{{end}}
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

// Create creates a new {{ structureName }}.
{{ if (hasID) }} func (t *{{ storageName | lowerCamelCase }}) Create(ctx context.Context, model *{{structureName}}) (*{{IDType}}, error) { {{ else }} func (t *{{ storageName | lowerCamelCase }}) Create(ctx context.Context, model *{{structureName}}) error { {{ end }}
	if model == nil {
		{{ if (hasID) }}return nil, errors.New("model is nil") {{ else }}return errors.New("model is nil") {{ end }}
	}

	{{- range $index, $field := fields }}
	{{- if not ($field | isRelation) }}
	{{- if ($field | isRepeated) }}
	// get value of {{ $field | fieldName | lowerCamelCase }}
	{{ $field | fieldName | lowerCamelCase }}, err := model.{{ $field | fieldName }}.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get {{ $field | fieldName | lowerCamelCase }} value: %w", err)
	}
	{{- end}}
	{{- end}}
	{{- end}}

	query := t.queryBuilder.Insert("{{ tableName }}").
		Columns(
			{{- range $index, $field := fields }}
			{{- if not ($field | isRelation) }}
			{{- if not ($field | isAutoIncrement ) }}
			{{- if not ($field | isDefaultUUID ) }}
			"{{ $field | sourceName }}",
			{{- end}}
			{{- end}}
			{{- end}}
			{{- end}}
		).
		Values(
			{{- range $index, $field := fields }}
			{{- if not ($field | isRelation) }}
			{{- if not ($field | isAutoIncrement ) }}
			{{- if not ($field | isDefaultUUID ) }}
			{{- if ($field | isRepeated) }}
			{{ $field | fieldName | lowerCamelCase }},
			{{- else }}
			model.{{ $field | fieldName }},
			{{- end}}
			{{- end}}
			{{- end}}
			{{- end}}
			{{- end}}
		)
	{{ if (hasID) }}
		// add RETURNING "id" to query
		query = query.Suffix("RETURNING \"id\"")
	{{ end }}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		{{ if (hasID) }} return nil, fmt.Errorf("failed to build query: %w", err) {{ else }} return fmt.Errorf("failed to build query: %w", err) {{ end }}
	}

	{{ if (hasID) }}var id {{IDType}}
	err = t.DB(ctx).QueryRow(sqlQuery, args...).Scan(&id) {{ else }} _, err = t.DB(ctx).Exec(sqlQuery, args...) {{ end }}
	if err != nil {
		if IsPgUniqueViolation(err) {
			{{ if (hasID) }}return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ else }}return errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ end }}
		}

		{{ if (hasID) }} return nil, fmt.Errorf("failed to create {{ structureName }}: %w", err) {{ else }} return fmt.Errorf("failed to create {{ structureName }}: %w", err) {{ end }}
	}

	{{ if (hasID) }}
	{{- range $index, $field := fields }}
	{{- if and ($field | isRelation) ($field | relationAllowSubCreating) }}
	    if model.{{ $field | fieldName }} != nil { {{ if ($field | isRepeated) }}
			for _, item := range model.{{ $field | fieldName }} {
				item.{{ $field | relationIDFieldName }} = id
				s := New{{ $field | relationStorageName }}(t.db)
                {{ if ($field | hasIDFromRelation) }} _, err := s.Create(ctx, item) {{ else }} err := s.Create(ctx, item) {{ end }}
				if err != nil {
				{{ if (hasID) }} return nil, fmt.Errorf("failed to create {{ $field | fieldName }}: %w", err) {{ else }} return fmt.Errorf("failed to create {{ structureName }}: %w", err) {{ end }}
				}
			} {{ else }}
			s := New{{ $field | relationStorageName }}(t.db)
			model.{{ $field | fieldName }}.{{ $field | relationIDFieldName }} = id
			{{ if ($field | hasIDFromRelation) }} _, err := s.Create(ctx, model.{{ $field | fieldName }}) {{ else }} err := s.Create(ctx, model.{{ $field | fieldName }}) {{ end }}
			if err != nil {
				{{ if (hasID) }} return nil, fmt.Errorf("failed to create {{ $field | fieldName }}: %w", err) {{ else }} return fmt.Errorf("failed to create {{ structureName }}: %w", err) {{ end }}
			} {{- end}}
	    } {{- end}}
	{{- end}}
	{{- end}}

	{{ if (hasID) }} return &id, nil {{ else }} return nil {{ end }}
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
