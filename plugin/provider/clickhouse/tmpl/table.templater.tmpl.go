package tmpl

const TableTemplate = `
{{ template "storage" . }}
{{ template "structure" . }}
{{ template "table_conditions" . }}
{{ template "create_method" . }}
{{ template "batch_create_method" . }}
{{ template "update_method" . }}
{{ template "delete_method" . }}
{{- if (hasPrimaryKey) }}
{{ template "get_by_id_method" . }}
{{- end }}
{{ template "find_many_method" . }}
{{ template "find_one_method" . }}
{{ template "count_method" . }}
{{ template "find_with_pagination" . }}
{{ template "lock_method" . }}
{{ template "raw_method" . }}
`

const TableConditionFilters = `
{{ range $key, $fieldMess := messages_for_filter }}
	{{- if len $fieldMess.GetField }}
		// {{ $fieldMess.GetName | camelCase }}Filters is a struct that holds filters for {{ $fieldMess.GetName }}.
		type {{structureName}}Filters struct {
			{{ range $field := $fieldMess.GetField }}
			{{- if not ($field | isRelation) }}
				{{- if (findPointer $field) }}
					{{ $field | fieldName }} {{ $field | fieldType }}
				{{- else }}
					{{ $field | fieldName }} *{{ $field | fieldType }}
				{{- end }}
			{{- end }}
			{{- end }}
		}
	{{- end }}
{{ end }}


{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Eq returns a condition that checks if the field equals the value.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Eq(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return EqualsCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotEq returns a condition that checks if the field equals the value.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotEq(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return NotEqualsCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GT greaterThanCondition than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GT(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return GreaterThanCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LT less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LT(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return LessThanCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GTE greater than or equal condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GTE(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return GreaterThanOrEqualCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LTE less than or equal condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LTE(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return LessThanOrEqualCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
   {{ range $field := $fieldMess.GetField }}
	{{- if not ($field | isRelation) }}
	{{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Between between condition.
	func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Between(min, max {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
		return BetweenCondition{Field: "{{ $field.GetName }}", Min: min, Max: max}
	}
	{{ end }}
	{{ end }}
	{{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
   {{- if ($field | isCurrentOptional) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNull checks if the {{ $field.GetName }} is NULL.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNull() FilterApplier {
      return IsNullCondition{Field: "{{ $field.GetName }}"}
    }

	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNotNull checks if the {{ $field.GetName }} is NOT NULL.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNotNull() FilterApplier {
      return IsNotNullCondition{Field: "{{ $field.GetName }}"}
    }
   {{ end }}
   {{ end }}
   {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
   {{- if ($field | isValidLike) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}ILike iLike condition %
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}ILike(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return ILikeCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
   {{- if ($field | isValidLike) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Like like condition %
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Like(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return LikeCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
   {{- if ($field | isValidLike) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotLike not like condition
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotLike(value {{- if (findPointer $field) }} {{ $field | fieldTypeWP }} {{- else }} {{ $field | fieldType }} {{- end }}) FilterApplier {
      return NotLikeCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
   {{- if ($field | isValidNull) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNull is null condition 
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNull() FilterApplier {
      return IsNullCondition{Field: "{{ $field.GetName }}"}
    }
  {{ end }}
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
   {{- if ($field | isValidNull) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNotNull is not null condition
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNotNull() FilterApplier {
      return IsNotNullCondition{Field: "{{ $field.GetName }}"}
    }
   {{ end }}
   {{ end }}
   {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}In condition
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}In(values ...interface{}) FilterApplier {
      return InCondition{Field: "{{ $field.GetName }}", Values: values}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotIn not in condition
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotIn(values ...interface{}) FilterApplier {
      return NotInCondition{Field: "{{ $field.GetName }}", Values: values}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

{{ range $key, $fieldMess := messages_for_filter }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}OrderBy sorts the result in ascending order.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}OrderBy(asc bool) FilterApplier {
      return OrderBy("{{ $field.GetName }}", asc)
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}
`

const TableFindWithPaginationMethodTemplate = `
// FindManyWithCursorPagination finds multiple {{ structureName }} using cursor-based pagination.
func (t *{{ storageName | lowerCamelCase }}) FindManyWithCursorPagination(
	ctx context.Context,
	limit int,
	cursor *string,
	cursorProvider CursorProvider,
	builders ...*QueryBuilder,
) ([]*{{structureName}}, *CursorPaginator, error) {
	if limit <= 0 {
		limit = 10
	}

	if cursorProvider == nil {
		return nil, nil, errors.New("cursor provider is required")
	}

	if cursor != nil && *cursor != "" {
		builders = append(builders, cursorProvider.CursorBuilder(*cursor))
	}

	builders = append(builders, LimitBuilder(uint64(limit+1)))
	records, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to find {{ structureName }}")
	}

	var nextCursor *string
	if len(records) > limit {
		lastRecord := records[limit]
		records = records[:limit]
		nextCursor = cursorProvider.GetCursor(lastRecord)
	}

	paginator := &CursorPaginator{
		Limit:     limit,
		NextCursor: nextCursor,
	}

	return records, paginator, nil
}
`

const TableLockMethodTemplate = `
// clickhouse does not support row-level locking.
`

const TableCountMethodTemplate = `
`

const TableFindOneMethodTemplate = `
// FindOne finds a single {{ structureName }} based on the provided options.
func (t *{{ storageName | lowerCamelCase }}) FindOne(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to findOne {{ structureName }}")
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}
`

const TableFindManyMethodTemplate = `
// FindMany finds multiple {{ structureName }} based on the provided options.
func (t *{{ storageName | lowerCamelCase }}) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*{{structureName}}, error) {
	// build query
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName())

	// set default options
	options := &Options{}

 	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}

		// apply custom filters
		query = builder.ApplyCustomFilters(query)

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
		for _, option := range builder.sortOptions {
			query = option.Apply(query)
		}

	    // apply options
		for _, o := range builder.options {
			o(options)
		}
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB().QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()
	
	var results []*{{structureName}}
	for rows.Next() {
		model := &{{structureName}}{}
		if err := model.ScanRows(rows); err != nil {
			return nil, errors.Wrap(err, "failed to scan {{ structureName }}")
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over rows")
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
		return nil, errors.Wrap(err, "find one {{ structureName }}: ")
	}

	return model, nil
}
`

const TableRawQueryMethodTemplate = `
// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *{{ storageName | lowerCamelCase }}) Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.DB().ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *{{ storageName | lowerCamelCase }}) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.DB().QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *{{ storageName | lowerCamelCase }}) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB().QueryContext(ctx, query, args...)
}
`

const TableDeleteMethodTemplate = ``

const TableUpdateMethodTemplate = ``

const StructureTemplate = `
// {{ structureName }} is a struct for the "{{ tableName }}" table.
type {{ structureName }} struct {
{{ range $field := fields }}
	{{ $field | fieldName }} {{ $field | fieldType }}{{if not ($field | isRelation) }}` + "" + `{{end}}{{end}}
}

// TableName returns the table name.
func (t *{{ structureName }}) TableName() string {
	return "{{ tableName }}"
}

// ScanRow scans a row into a {{ structureName }}.
func (t *{{ structureName }}) ScanRow(r *sql.Row) error {
	return r.Scan({{ range $field := fields }} {{if not ($field | isRelation) }} &t.{{ $field | fieldName }}, {{ end }}{{ end }})
}

// ScanRows scans a single row into the {{ structureName }}.
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

const TableBatchCreateMethodTemplate = `
// BatchCreate creates multiple {{ structureName }} records in a single batch.
func (t *{{ storageName | lowerCamelCase }}) BatchCreate(ctx context.Context, models []*{{structureName}}, opts ...Option) error {
	if len(models) == 0 {
		return errors.New("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return errors.New("relations are not supported in batch create")
	}

	query := t.queryBuilder.Insert(t.TableName()).
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
		)

	for _, model := range models {
		if model == nil {
			return errors.New("one of the models is nil")
		}
		query = query.Values(
			{{- range $index, $field := fields }}
			{{- if not ($field | isRelation) }}
			{{- if not ($field | isAutoIncrement ) }}
			{{- if not ($field | isDefaultUUID ) }}

			{{- if ($field | isRepeated) }}
				{{ $field | fieldName | lowerCamelCase }},
			{{- else }}
				{{- if (findPointer $field) }}
				nullValue(model.{{ $field | fieldName }}),
				{{- else }}
				model.{{ $field | fieldName }},
				{{- end }}
			{{- end}}

			{{- end}}
			{{- end}}
			{{- end}}
			{{- end}}
		)
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB().QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to execute bulk insert")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "rows iteration error")
	}

	return nil
}
`

const TableCreateMethodTemplate = `
// Create creates a new {{ structureName }}.
func (t *{{ storageName | lowerCamelCase }}) Create(ctx context.Context, model *{{structureName}}, opts ...Option) error { 
	if model == nil {
		return errors.New("model is nil")
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
		return errors.Wrap(err, "failed to get value of {{ $field | fieldName }}")
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
			
				{{- if (findPointer $field) }}
				nullValue(model.{{ $field | fieldName }}),
				{{- else }}
				model.{{ $field | fieldName }},
				{{- end }}

			{{- end}}

			{{- end}}
			{{- end}}
			{{- end}}
			{{- end}}
	)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB().ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to create {{ structureName }}")
	}

	{{- range $index, $field := fields }}
	{{- if and ($field | isRelation) ($field | relationAllowSubCreating) }}
	    if options.relations && model.{{ $field | fieldName }} != nil { {{ if ($field | isRepeated) }}
			for _, item := range model.{{ $field | fieldName }} {
				s, err := New{{ $field | relationStorageName }}(t.config)
				if err != nil {
					return errors.Wrap(err, "failed to create {{ $field | fieldName }}")
				}

				err = s.Create(ctx, item)
				if err != nil {
					return errors.Wrap(err, "failed to create {{ $field | fieldName }}")
				}
			} {{ else }}
			s, err := New{{ $field | relationStorageName }}(t.config)
			if err != nil {
				return errors.Wrap(err, "failed to create {{ $field | fieldName }}")
			}

			err = s.Create(ctx, model.{{ $field | fieldName }})
			if err != nil {
				return errors.Wrap(err, "failed to create {{ $field | fieldName }}")
			} {{- end}}
	    } {{- end}}
	{{- end}}

	return nil
}
`

const TableStorageTemplate = `
// {{ storageName | lowerCamelCase }} is a struct for the "{{ tableName }}" table.
type {{ storageName | lowerCamelCase }} struct {
	config *Config
	queryBuilder sq.StatementBuilderType
}

{{ if .CRUDSchemas }}
// {{structureName}}TableManager is an interface for managing the {{ tableName }} table.
type {{structureName}}TableManager interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
}
{{ end }}

// {{structureName}}CRUDOperations is an interface for managing the {{ tableName }} table.
type {{structureName}}CRUDOperations interface {
	Create(ctx context.Context, model *{{structureName}}, opts ...Option) error
	BatchCreate(ctx context.Context, models []*{{structureName}}, opts ...Option) error
	{{- if (hasPrimaryKey) }}
	FindBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, id {{IDType}}, opts ...Option) (*{{ structureName }}, error)
	{{- end }}
}

// {{structureName}}SearchOperations is an interface for searching the {{ tableName }} table.
type {{structureName}}SearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*{{structureName}}, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error)
}

// {{structureName}}PaginationOperations is an interface for pagination operations.
type {{structureName}}PaginationOperations interface {
	FindManyWithCursorPagination(ctx context.Context, limit int, cursor *string, cursorProvider CursorProvider, builders ...*QueryBuilder) ([]*{{structureName}}, *CursorPaginator, error)
}

// {{structureName}}RelationLoading is an interface for loading relations.
type {{structureName}}RelationLoading interface {
	{{- range $index, $field := fields }}
	{{- if and ($field | isRelation) }}
	Load{{ $field | pluralFieldName }} (ctx context.Context, model *{{structureName}}, builders ...*QueryBuilder) error
	{{- end }}
	{{- end }}
	{{- range $index, $field := fields }}
	{{- if and ($field | isRelation) }}
	LoadBatch{{ $field | pluralFieldName }} (ctx context.Context, items []*{{structureName}}, builders ...*QueryBuilder) error
	{{- end }}
	{{- end }}
}

// {{structureName}}RawQueryOperations is an interface for executing raw queries.
type {{structureName}}RawQueryOperations interface {
	Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// {{ storageName }} is a struct for the "{{ tableName }}" table.
type {{ storageName }} interface {
{{ if .CRUDSchemas }}
    {{structureName}}TableManager
{{ end }}
	{{structureName}}CRUDOperations
	{{structureName}}SearchOperations
	{{structureName}}PaginationOperations
	{{structureName}}RelationLoading
	{{structureName}}RawQueryOperations
}

// New{{ storageName }} returns a new {{ storageName | lowerCamelCase }}.
func New{{ storageName }}(config *Config) ({{ storageName }}, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.DB == nil {
		return nil, errors.New("config.DB connection is nil")
	}

	return &{{ storageName | lowerCamelCase }}{
		config: config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *{{ storageName | lowerCamelCase }}) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *{{ storageName | lowerCamelCase }}) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
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

// DB returns the underlying DB. This is useful for doing transactions.
func (t *{{ storageName | lowerCamelCase }}) DB() QueryExecer {
	return t.config.DB
}

{{ if .CRUDSchemas }}
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
		{{ $field | sourceName }} {{if ($field | isAutoIncrement) }} SERIAL{{else}}{{ $field | postgresType }}{{end}}{{if $field | isPrimaryKey }} PRIMARY KEY{{end}}{{ if and (isNotNull $field) (not (isAutoIncrement $field)) }} NOT NULL{{ end }}{{if ($field | getDefaultValue) }} DEFAULT {{$field | getDefaultValue}}{{end}}{{if not ( $field | isLastField )}},{{end}}
		{{- end}}
		{{- end}});
		-- Other entities
		{{- if (comment) }}
		COMMENT ON TABLE {{ tableName }} IS '{{ comment }}';
		{{- end}}
		{{- range $index, $field := fields }}
		{{- if ($field | hasUnique) }}
		CREATE UNIQUE INDEX IF NOT EXISTS {{ tableName }}_{{ $field | sourceName }}_unique_idx ON {{ tableName }} USING btree ({{ $field | sourceName }});
		{{- end}}
		{{- end}}

		{{- range $index, $fields := getStructureUniqueIndexes }}
		CREATE UNIQUE INDEX IF NOT EXISTS {{ tableName }}_unique_idx_{{ $fields | sliceToString }} ON {{ tableName }} USING btree (
        {{- $length := sub (len $fields) 1 }}
        {{- range $i, $field := $fields }}
            {{ $field | sourceName }}{{ if lt $i $length }}, {{ end }}
        {{- end }}
    	);
		{{- end }}


		{{- range $index, $field := fields }}
		{{- if ($field | hasIndex) }}
		CREATE INDEX IF NOT EXISTS {{ tableName }}_{{ $field | sourceName }}_idx ON {{ tableName }} USING btree ({{ $field | sourceName }});
		{{- end}}
		{{- end}}
		{{- range $index, $field := fields }}
		{{- if ($field | isRelation) }}
		{{- if ($field | isForeign) }}
		-- Foreign keys for {{ $field | relationTableName }}
		ALTER TABLE {{ tableName }}
		ADD FOREIGN KEY ({{ $field | getFieldSource }}) REFERENCES {{ $field | relationTableName }}({{ $field | getRefSource }})
		{{- if ($field | isCascade) }}
		ON DELETE CASCADE;
		{{- else }}; 
        {{- end}}
		{{- end}}
		{{- end}}
		{{- end }}
	` + "`" + `

	_, err := t.DB().ExecContext(ctx,sqlQuery)
	return err
}

// DropTable drops the table.
func (t *{{ storageName | lowerCamelCase }}) DropTable(ctx context.Context) error {
	sqlQuery := ` + "`" + `
		DROP TABLE IF EXISTS {{ tableName }};
	` + "`" + `

	_, err := t.DB().ExecContext(ctx,sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *{{ storageName | lowerCamelCase }}) TruncateTable(ctx context.Context) error {
	sqlQuery := ` + "`" + `
		TRUNCATE TABLE {{ tableName }};
	` + "`" + `

	_, err := t.DB().ExecContext(ctx,sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method 
func (t *{{ storageName | lowerCamelCase }}) UpgradeTable(ctx context.Context) error {
	return nil
}
{{ end }}

{{- range $index, $field := fields }}
{{- if and ($field | isRelation) }}
// Load{{ $field | pluralFieldName }} loads the {{ $field | pluralFieldName }} relation.
func (t *{{ storageName | lowerCamelCase }}) Load{{ $field | pluralFieldName }}(ctx context.Context, model *{{structureName}}, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "{{structureName}} is nil")
	}

	// New{{ $field | relationStorageName }} creates a new {{ $field | relationStorageName }}.
	s, err := New{{ $field | relationStorageName }}(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create {{ $field | relationStorageName }}")
	}

	{{- if ($field | isOptional) }}
		// Check if the optional field is nil
		if model.{{ $field | getFieldID }} == nil {
			// If nil, do not attempt to load the relation
			return nil
		}
		// Add the filter for the relation with dereferenced value
		builders = append(builders, FilterBuilder({{ $field | relationStructureName }}{{ $field | getRefID }}Eq(*model.{{ $field | getFieldID }})))
	{{- else }}
		// Add the filter for the relation without dereferencing
		builders = append(builders, FilterBuilder({{ $field | relationStructureName }}{{ $field | getRefID }}Eq(model.{{ $field | getFieldID }})))
	{{- end }}

	{{- if ($field | isRepeated) }}
		relationModels, err := s.FindMany(ctx, builders...)
		if err != nil {
			return errors.Wrap(err, "failed to find many {{ $field | relationStorageName }}")
		}

		model.{{ $field | fieldName }} = relationModels
	{{- else }}
		relationModel, err := s.FindOne(ctx, builders...)
		if err != nil {
			return errors.Wrap(err, "failed to find one {{ $field | relationStorageName }}")
		}

		model.{{ $field | fieldName }} = relationModel
	{{- end }}
	return nil
}
{{- end }}
{{- end }}

{{- range $index, $field := fields }}
{{- if and ($field | isRelation) }}
// LoadBatch{{ $field | pluralFieldName }} loads the {{ $field | pluralFieldName }} relation.
func (t *{{ storageName | lowerCamelCase }}) LoadBatch{{ $field | pluralFieldName }}(ctx context.Context, items []*{{structureName}}, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		{{- if ($field | isOptional) }}
			// Check if the field is nil for optional fields
			if item.{{ $field | getFieldID }} == nil {
				// Skip nil values for optional fields
				continue
			}
			// Append dereferenced value for optional fields
			requestItems = append(requestItems, *item.{{ $field | getFieldID }})
		{{- else }}
			// Append the value directly for non-optional fields
			requestItems = append(requestItems, item.{{ $field | getFieldID }})
		{{- end }}
	}

	// New{{ $field | relationStorageName }} creates a new {{ $field | relationStorageName }}.
	s, err := New{{ $field | relationStorageName }}(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create {{ $field | relationStorageName }}")
	}

	// Add the filter for the relation
	{{- if ($field | isOptional) }}
		// Ensure that requestItems are not empty before adding the builder
		if len(requestItems) > 0 {
			builders = append(builders, FilterBuilder({{ $field | relationStructureName }}{{ $field | getRefID }}In(requestItems...)))
		}
	{{- else }}
		builders = append(builders, FilterBuilder({{ $field | relationStructureName }}{{ $field | getRefID }}In(requestItems...)))
	{{- end }}

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many {{ $field | relationStorageName }}")
	}

	{{- if ($field | isRepeated) }}
	resultMap := make(map[interface{}][]*{{ $field | relationStructureName }})
	{{- else }}
	resultMap := make(map[interface{}]*{{ $field | relationStructureName }})
	{{- end }}
	for _, result := range results {
		{{- if ($field | isRepeated) }}
		resultMap[result.{{ $field | getRefID }}] = append(resultMap[result.{{ $field | getRefID }}], result)
		{{- else }}
		resultMap[result.{{ $field | getRefID }}] = result
		{{- end }}
	}

	// Assign {{ $field | relationStructureName }} to items
	for _, item := range items {
		{{- if ($field | isOptional) }}
			// Skip assignment if the field is nil
			if item.{{ $field | getFieldID }} == nil {
				continue
			}
			// Assign the relation if it exists in the resultMap
			if v, ok := resultMap[*item.{{ $field | getFieldID }}]; ok {
				item.{{ $field | fieldName }} = v
			}
		{{- else }}
			// Assign the relation directly for non-optional fields
			if v, ok := resultMap[item.{{ $field | getFieldID }}]; ok {
				item.{{ $field | fieldName }} = v
			}
		{{- end }}
	}

	return nil
}
{{- end }}
{{- end }}
`
