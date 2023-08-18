package tmpl

const TableTemplate = `
{{ template "storage" . }}
{{ template "structure" . }}
{{ template "create_method" . }}
{{ template "update_method" . }}
{{- if (hasPrimaryKey) }}
{{ template "delete_method" . }}
{{- end }}
{{- if (hasPrimaryKey) }}
{{ template "get_by_id_method" . }}
{{- end }}
{{ template "find_many_method" . }}
{{ template "find_one_method" . }}
`

const TableFindOneMethodTemplate = `
// findOne finds a single {{ structureName }} based on the provided options.
func (t *{{ storageName | lowerCamelCase }}) FindOne(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no {{ structureName }} found")
	}

	return results[0], nil
}
`

const TableFindManyMethodTemplate = `
// findMany finds multiple {{ structureName }} based on the provided options.
func (t *{{ storageName | lowerCamelCase }}) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*{{structureName}}, error) {
	// build query
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName())

 	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}
		// apply pagination
		if builder.pagination != nil {
			if builder.pagination.limit != nil {
				query = query.Limit(*builder.pagination.limit)
			}
			if builder.pagination.offset != nil {
				query = query.Offset(*builder.pagination.offset)
			}
		}
		// apply sorting
		// ...
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := t.DB(ctx).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find {{ structureName }}: %w", err)
	}
	defer rows.Close()
	
	var results []*{{structureName}}
	for rows.Next() {
		model := &{{structureName}}{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan {{ structureName }}: %w", err)
		}
		results = append(results, model)
	}
	
	return results, nil
}
`

const TableGetByIDMethodTemplate = `
// FindBy{{ getPrimaryKey.GetName | camelCase }} retrieves a {{ structureName }} by its {{ getPrimaryKey.GetName }}.
func (t *{{ storageName | lowerCamelCase }}) FindBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, id {{IDType}}, opts ...Option) (*{{ structureName }}, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter({{ messageName }}{{ getPrimaryKey.GetName | camelCase }}Eq(id))
		builder.WithOptions(opts...)
	}
	
	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to get {{ structureName }}: %w", err)
	}

	return model, nil
}
`

const TableDeleteMethodTemplate = `
// Delete removes an existing {{ structureName }} by its ID.
func (t *{{ storageName | lowerCamelCase }}) DeleteBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, {{getPrimaryKey.GetName}} {{IDType}}, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("{{ tableName }}").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx,sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete {{ structureName }}: %w", err)
	}

	return nil
}
`

const TableUpdateMethodTemplate = `
// {{ structureName }}Update is used to update an existing {{ structureName }}.
type {{ structureName }}Update struct {
	{{- range $index, $field := fields }}
	{{- if not ($field | isRelation) }}
	{{- if not ($field | isAutoIncrement) }}
	{{- if not ($field | isPrimary) }}
	{{ $field | fieldName }} {{- if not ($field | findPointer) }}*{{- end }}{{ $field | fieldType }}
	{{- end}}
	{{- end}}
	{{- end}}
	{{- end}}
}

// Update updates an existing {{ structureName }} based on non-nil fields.
func (t *{{ storageName | lowerCamelCase }}) Update(ctx context.Context, id {{IDType}}, updateData *{{structureName}}Update) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("{{ tableName }}")

	{{- range $index, $field := fields }}
	{{- if not ($field | isRelation) }}
	{{- if not ($field | isAutoIncrement) }}
	{{- if not ($field | isPrimary) }}
	if updateData.{{ $field | fieldName }} != nil {
		{{- if ($field | isRepeated) }}
		value, err := updateData.{{ $field | fieldName }}.Value()
		if err != nil {
			return fmt.Errorf("failed to get {{ $field | fieldName | lowerCamelCase }} value: %w", err)
		}
		query = query.Set("{{ $field | sourceName }}", value)
		{{- else }}
		query = query.Set("{{ $field | sourceName }}", *updateData.{{ $field | fieldName }})
		{{- end}}
	}
	{{- end}}
	{{- end}}
	{{- end}}
	{{- end}}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx,sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update {{ structureName }}: %w", err)
	}

	return nil
}
`

const TableCreateMethodTemplate = `
// Create creates a new {{ structureName }}.
{{ if (hasID) }} func (t *{{ storageName | lowerCamelCase }}) Create(ctx context.Context, model *{{structureName}}, opts ...Option) (*{{IDType}}, error) { {{ else }} func (t *{{ storageName | lowerCamelCase }}) Create(ctx context.Context, model *{{structureName}}, opts ...Option) error { {{ end }}
	if model == nil {
		{{ if (hasID) }}return nil, errors.New("model is nil") {{ else }}return errors.New("model is nil") {{ end }}
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
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
	err = t.DB(ctx).QueryRowContext(ctx,sqlQuery, args...).Scan(&id) {{ else }} _, err = t.DB(ctx).ExecContext(ctx,sqlQuery, args...) {{ end }}
	if err != nil {
		if IsPgUniqueViolation(err) {
			{{ if (hasID) }}return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ else }}return errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ end }}
		}

		{{ if (hasID) }} return nil, fmt.Errorf("failed to create {{ structureName }}: %w", err) {{ else }} return fmt.Errorf("failed to create {{ structureName }}: %w", err) {{ end }}
	}

	{{ if (hasID) }}
	{{- range $index, $field := fields }}
	{{- if and ($field | isRelation) ($field | relationAllowSubCreating) }}
	    if options.relations && model.{{ $field | fieldName }} != nil { {{ if ($field | isRepeated) }}
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

const TableStorageTemplate = `
// {{ storageName | lowerCamelCase }} is a struct for the "{{ tableName }}" table.
type {{ storageName | lowerCamelCase }} struct {
	db *sql.DB // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

type {{ storageName }} interface {
	// CreateTable creates the table.
	CreateTable(ctx context.Context) error
	// DropTable drops the table.
	DropTable(ctx context.Context) error
	// TruncateTable truncates the table.
	TruncateTable(ctx context.Context) error
	// UpgradeTable upgrades the table.
	UpgradeTable(ctx context.Context) error
	// Create creates a new {{ structureName }}.
	{{- if (hasID) }}
	Create(ctx context.Context, model *{{structureName}}, opts ...Option) (*{{IDType}}, error)
	{{- else }} 
	Create(ctx context.Context, model *{{structureName}}, opts ...Option) error
	{{- end }}
	// Update updates an existing {{ structureName }}.
	Update(ctx context.Context, id {{IDType}}, updateData *{{structureName}}Update) error

	{{- if (hasPrimaryKey) }}
	// Delete removes an existing {{ structureName }} by its ID.
	DeleteBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, {{getPrimaryKey.GetName}} {{IDType}}, opts ...Option) error
	{{- end }}
	{{- if (hasPrimaryKey) }}
	// FindBy{{ getPrimaryKey.GetName | camelCase }} retrieves a {{ structureName }} by its {{ getPrimaryKey.GetName }}.
	FindBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, id {{IDType}}, opts ...Option) (*{{ structureName }}, error)
	{{- end }}
	// FindMany finds multiple {{ structureName }} based on the provided options.
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*{{structureName}}, error)
	// FindOne finds a single {{ structureName }} based on the provided options.
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error)	
}

// New{{ storageName }} returns a new {{ storageName | lowerCamelCase }}.
func New{{ storageName }}(db *sql.DB) {{ storageName }} {
	return &{{ storageName | lowerCamelCase }}{
		db: db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *{{ storageName | lowerCamelCase }}) TableName() string {
	return "{{ tableName }}"
}

// Columns returns the columns for the table.
func (t *{{ storageName | lowerCamelCase }}) Columns() []string {
	return []string{
		{{ range $field := fields }}{{if not ($field | isRelation) }}"{{ $field | sourceName }}",{{ end }}{{ end }}
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
func (t *{{ storageName | lowerCamelCase }}) CreateTable(ctx context.Context) error {
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

	_, err := t.db.ExecContext(ctx,sqlQuery)
	return err
}

// DropTable drops the table.
func (t *{{ storageName | lowerCamelCase }}) DropTable(ctx context.Context) error {
	sqlQuery := ` + "`" + `
		DROP TABLE IF EXISTS {{ tableName }};
	` + "`" + `

	_, err := t.db.ExecContext(ctx,sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *{{ storageName | lowerCamelCase }}) TruncateTable(ctx context.Context) error {
	sqlQuery := ` + "`" + `
		TRUNCATE TABLE {{ tableName }};
	` + "`" + `

	_, err := t.db.ExecContext(ctx,sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
func (t *{{ storageName | lowerCamelCase }}) UpgradeTable(ctx context.Context) error {
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

// ScanRow scans a single row into the {{ structureName }}.
func (t *{{ structureName }}) ScanRows(r *sql.Rows) error {
	return r.Scan(
		{{- range $index, $field := fields }}
		{{- if not ($field | isRelation) }}
		&t.{{ $field | fieldName }},
		{{- end}}
		{{- end }}
	)
}
`
