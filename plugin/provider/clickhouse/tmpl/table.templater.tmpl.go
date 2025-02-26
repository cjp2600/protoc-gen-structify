package tmpl

const TableTemplate = `
{{ template "storage" . }}
{{ template "structure" . }}
{{ template "table_conditions" . }}
{{ template "async_create_method" . }}
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

const TableFindWithPaginationMethodTemplate = ``

const TableLockMethodTemplate = ``

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

		// apply custom table name
		query = builder.ApplyCustomTableName(query)

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

	rows, err := t.DB().Query(ctx, sqlQuery, args...)
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
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
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

const TableGetByIDMethodTemplate = ``

const TableRawQueryMethodTemplate = `
// Select executes a raw query and returns the result.
func (t *{{ storageName | lowerCamelCase }}) Select(ctx context.Context, query string, dest any, args ...any) error {
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *{{ storageName | lowerCamelCase }}) Exec(ctx context.Context, query string, args ...interface{}) error {
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *{{ storageName | lowerCamelCase }}) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *{{ storageName | lowerCamelCase }}) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *{{ storageName | lowerCamelCase }}) Conn() driver.Conn {
	return t.DB()
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
func (t *{{ structureName }}) ScanRow(row driver.Row) error {
	return row.Scan(
		{{- range $field := fields }}
		{{- if not ($field | isRelation) }}
		&t.{{ $field | fieldName }},
		{{- end }}
		{{- end }}
	)
}`

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

	batch, err := t.DB().PrepareBatch(ctx, "INSERT INTO " + t.TableName())
	if err != nil {
		return errors.Wrap(err, "failed to prepare batch")
	}

	for _, model := range models {
		if model == nil {
			return errors.New("one of the models is nil")
		}

		{{- range $index, $field := fields }}
		{{- if not ($field | isRelation) }}
		{{- if ($field | isRepeated) }}
		// Get value of {{ $field | fieldName | lowerCamelCase }}
		{{ $field | fieldName | lowerCamelCase }}, err := model.{{ $field | fieldName }}.Value()
		if err != nil {
			return errors.Wrap(err, "failed to get value of {{ $field | fieldName }}")
		}
		{{- end}}
		{{- end}}
		{{- end}}

		{{ if isHasRepeated }}err = batch.Append({{ else }}err := batch.Append({{ end }}
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
		if err != nil {
			return errors.Wrap(err, "failed to append to batch")
		}
	}

	if err := batch.Send(); err != nil {
		return errors.Wrap(err, "failed to execute batch insert")
	}

	return nil
}`

const TableCreateAsyncMethodTemplate = `
// AsyncCreate asynchronously inserts a new {{ structureName }}.
func (t *{{ storageName | lowerCamelCase }}) AsyncCreate(ctx context.Context, model *{{structureName}}, opts ...Option) error { 
	if model == nil {
		return errors.New("model is nil")
	}

	// Set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	{{- range $index, $field := fields }}
	{{- if not ($field | isRelation) }}
	{{- if ($field | isRepeated) }}
	// Get value of {{ $field | fieldName | lowerCamelCase }}
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

	if err := t.DB().AsyncInsert(ctx, sqlQuery, false, args...); err != nil {
		return errors.Wrap(err, "failed to asynchronously create {{ structureName }}")
	}

	{{- range $index, $field := fields }}
	{{- if and ($field | isRelation) ($field | relationAllowSubCreating) }}
	if options.relations && model.{{ $field | fieldName }} != nil { {{ if ($field | isRepeated) }}
		for _, item := range model.{{ $field | fieldName }} {
			s, err := New{{ $field | relationStorageName }}(t.config)
			if err != nil {
				return errors.Wrap(err, "failed to create {{ $field | fieldName }}")
			}

			err = s.AsyncCreate(ctx, item)
			if err != nil {
				return errors.Wrap(err, "failed to asynchronously create {{ $field | fieldName }}")
			}
		}
	{{- else }}
		s, err := New{{ $field | relationStorageName }}(t.config)
		if err != nil {
			return errors.Wrap(err, "failed to create {{ $field | fieldName }}")
		}

		err = s.AsyncCreate(ctx, model.{{ $field | fieldName }})
		if err != nil {
			return errors.Wrap(err, "failed to asynchronously create {{ $field | fieldName }}")
		}
	{{- end}}
	} {{- end}}
	{{- end}}

	return nil
}`

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

	err = t.DB().Exec(ctx, sqlQuery, args...)
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

// {{structureName}}CRUDOperations is an interface for managing the {{ tableName }} table.
type {{structureName}}CRUDOperations interface {
	Create(ctx context.Context, model *{{structureName}}, opts ...Option) error
	AsyncCreate(ctx context.Context, model *{{structureName}}, opts ...Option) error
	BatchCreate(ctx context.Context, models []*{{structureName}}, opts ...Option) error
}

// {{structureName}}SearchOperations is an interface for searching the {{ tableName }} table.
type {{structureName}}SearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*{{structureName}}, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error)
}

type {{structureName}}Settings interface {
	Conn() driver.Conn
	SetConfig(config *Config) {{ storageName }}
	SetQueryBuilder(builder sq.StatementBuilderType) {{ storageName }}
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
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// {{ storageName }} is a struct for the "{{ tableName }}" table.
type {{ storageName }} interface {
	{{structureName}}CRUDOperations
	{{structureName}}SearchOperations
	{{structureName}}RelationLoading
	{{structureName}}RawQueryOperations
	{{structureName}}Settings
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

func (t *{{ storageName | lowerCamelCase }}) SetConfig(config *Config) {{ storageName }} {
	t.config = config
	return t
}

func (t *{{ storageName | lowerCamelCase }}) SetQueryBuilder(builder sq.StatementBuilderType) {{ storageName }} {
	t.queryBuilder = builder
	return t
}

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
