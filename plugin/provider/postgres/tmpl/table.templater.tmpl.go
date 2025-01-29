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
			{{ $field | fieldName }} *{{ $field | fieldType }}
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Eq(value {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotEq(value {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GT(value {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LT(value {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GTE(value {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LTE(value {{ $field | fieldType }}) FilterApplier {
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
	func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Between(min, max {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}ILike(value {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Like(value {{ $field | fieldType }}) FilterApplier {
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
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotLike(value {{ $field | fieldType }}) FilterApplier {
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
// FindManyWithPagination finds multiple {{ structureName }} with pagination support.
func (t *{{ storageName | lowerCamelCase }}) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*{{structureName}}, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to count {{ structureName }}")
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Build the pagination object
	paginator := &Paginator{
		TotalCount: totalCount,
		Limit:      limit,
		Page:       page,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
	}

	// Add pagination to query builder
	builders = append(builders, PaginateBuilder(uint64(limit), uint64(offset)))

	// Find records using FindMany
	records, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to find {{ structureName }}")
	}

	return records, paginator, nil
}
`

const TableLockMethodTemplate = `
// SelectForUpdate lock locks the {{ structureName }} for the given ID.
func (t *{{ storageName | lowerCamelCase }}) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error) {
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName()).Suffix("FOR UPDATE")

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
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args)

	row := t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...)
	var model {{ structureName }}
    if err := model.ScanRow(row); err != nil {
        if errors.Is(err, sql.ErrNoRows){
            return nil, ErrRowNotFound
        }
        return nil, errors.Wrap(err, "failed to scan {{ structureName }}")
    }

	return &model, nil
}
`

const TableCountMethodTemplate = `
// Count counts {{ structureName }} based on the provided options.
func (t *{{ storageName | lowerCamelCase }}) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
	// build query
	query := t.queryBuilder.Select("COUNT(*)").From(t.TableName())

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
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args)

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "failed to scan count")
	}

	return count, nil
}
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
	t.logQuery(ctx, sqlQuery, args)

	rows, err := t.DB(ctx, false).QueryContext(ctx, sqlQuery, args...)
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
func (t *{{ storageName | lowerCamelCase }}) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *{{ storageName | lowerCamelCase }}) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *{{ storageName | lowerCamelCase }}) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
`

const TableDeleteMethodTemplate = `
{{- if (hasPrimaryKey) }}
// DeleteBy{{ getPrimaryKey.GetName | camelCase }} - deletes a {{ structureName }} by its {{ getPrimaryKey.GetName }}.
func (t *{{ storageName | lowerCamelCase }}) DeleteBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, {{getPrimaryKey.GetName}} {{IDType}}, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("{{ tableName }}").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args)

	_, err = t.DB(ctx, true).ExecContext(ctx,sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete {{ structureName }}")
	}

	return nil
}
{{- end }}

// DeleteMany removes entries from the {{ tableName }} table using the provided filters
func (t *{{ storageName | lowerCamelCase }}) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("{{ tableName }}")

	var withFilter bool
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.ApplyDelete(query)
			withFilter = true
		}
	}

	if !withFilter {
		return errors.New("filters are required for delete operation")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete {{ tableName }}")
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
	{{- if ($field | isCurrentOptional) }}
		// Use null types for optional fields
		{{ $field | fieldName }} {{ $field | fieldTypeToNullType }}
	{{- else }}
		// Use regular pointer types for non-optional fields
		{{ $field | fieldName }} {{- if not ($field | findPointer) }}*{{- end }}{{ $field | fieldType }}
	{{- end }}
	{{- end }}
	{{- end }}
	{{- end }}
	{{- end }}
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
	{{- if ($field | isCurrentOptional) }}
		// Handle fields that are optional and can be explicitly set to NULL
		if updateData.{{ $field | fieldName }}.Valid {
			{{- if (eq ($field | fieldTypeToNullType) "null.String") }}
			// Handle null.String specifically
			if updateData.{{ $field | fieldName }}.String == "" {
				query = query.Set("{{ $field | sourceName }}", nil) // Explicitly set NULL for empty string
			} else {
				query = query.Set("{{ $field | sourceName }}", updateData.{{ $field | fieldName }}.ValueOrZero())
			}
			{{- else if (eq ($field | fieldTypeToNullType) "null.Int") }}
			// Handle null.Int specifically
			query = query.Set("{{ $field | sourceName }}", updateData.{{ $field | fieldName }}.ValueOrZero())
			{{- else if (eq ($field | fieldTypeToNullType) "null.Float") }}
			// Handle null.Float specifically
			query = query.Set("{{ $field | sourceName }}", updateData.{{ $field | fieldName }}.ValueOrZero())
			{{- else if (eq ($field | fieldTypeToNullType) "null.Bool") }}
			// Handle null.Bool specifically
			query = query.Set("{{ $field | sourceName }}", updateData.{{ $field | fieldName }}.ValueOrZero())
			{{- else if (eq ($field | fieldTypeToNullType) "null.Time") }}
			// Handle null.Time specifically
			if updateData.{{ $field | fieldName }}.Time.IsZero() {
				query = query.Set("{{ $field | sourceName }}", nil) // Explicitly set NULL if time is zero
			} else {
				query = query.Set("{{ $field | sourceName }}", updateData.{{ $field | fieldName }}.Time)
			}
			{{- else if ($field | isJSON) }}
			if updateData.{{ $field | fieldName }}.Data == nil {
				query = query.Set("{{ $field | sourceName }}", nil) // Explicitly set NULL
			} else {
				query = query.Set("{{ $field | sourceName }}", updateData.{{ $field | fieldName }}.Data)
			}
			{{- else }}
			// Handle other null types
			// ... add more specific handling here if needed
			{{- end }}
		}
	{{- else }}
		// Handle fields that are not optional using a nil check
		if updateData.{{ $field | fieldName }} != nil {
			query = query.Set("{{ $field | sourceName }}", *updateData.{{ $field | fieldName }}) // Dereference pointer value
		}
	{{- end }}
	{{- end }}
	{{- end }}
	{{- end }}
	{{- end }}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to update {{ structureName }}")
	}

	return nil
}
`

const StructureTemplate = `
// {{ structureName }} is a struct for the "{{ tableName }}" table.
type {{ structureName }} struct {
{{ range $field := fields }}
	{{ $field | fieldName }} {{ $field | fieldType }}{{if not ($field | isRelation) }}` + " `db:\"{{ $field | sourceName }}\"`" + `{{end}}{{end}}
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
func (t *{{ storageName | lowerCamelCase }}) BatchCreate(ctx context.Context, models []*{{structureName}}, opts ...Option) ({{ if (hasID) }}[]string, {{ end }}error) {
	if len(models) == 0 {
		{{ if (hasID) }} return nil, errors.New("no models to insert") {{ else }} return errors.New("no models to insert") {{ end }}
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		{{ if (hasID) }} return nil, errors.New("relations are not supported in batch create") {{ else }} return errors.New("relations are not supported in batch create") {{ end }}
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
			{{ if (hasID) }} return nil, errors.New("one of the models is nil") {{ else }} return errors.New("one of the models is nil") {{ end }}
		}
		query = query.Values(
			{{- range $index, $field := fields }}
			{{- if not ($field | isRelation) }}
			{{- if not ($field | isAutoIncrement ) }}
			{{- if not ($field | isDefaultUUID ) }}
			model.{{ $field | fieldName }},
			{{- end}}
			{{- end}}
			{{- end}}
			{{- end}}
		)
	}

	{{ if (hasID) }}
	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT ("+options.ignoreConflictField+") DO NOTHING RETURNING \"id\"")
	} else {
		query = query.Suffix("RETURNING \"id\"")
	}
	{{ else }}
	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT ("+options.ignoreConflictField+") DO NOTHING")
	}
	{{ end }}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		{{ if (hasID) }} return nil, errors.Wrap(err, "failed to build query") {{ else }} return errors.Wrap(err, "failed to build query") {{ end }}
	}
	t.logQuery(ctx, sqlQuery, args)

	rows, err := t.DB(ctx, true).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			{{ if (hasID) }} return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ else }} return errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ end }}
		}
		{{ if (hasID) }} return nil, errors.Wrap(err, "failed to execute bulk insert") {{ else }} return errors.Wrap(err, "failed to execute bulk insert") {{ end }}
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	{{ if (hasID) }} var returnIDs []string {{ end }} {{ if (hasID) }}
	for rows.Next() {
		var {{ getPrimaryKey.GetName }} string
		if err := rows.Scan(&{{ getPrimaryKey.GetName }}); err != nil {
			{{ if (hasID) }} return nil, errors.Wrap(err, "failed to scan {{ getPrimaryKey.GetName }}") {{ else }} return errors.Wrap(err, "failed to scan {{ getPrimaryKey.GetName }}") {{ end }}
		}
		returnIDs = append(returnIDs, {{ getPrimaryKey.GetName }})
	}
	{{ end }}

	if err := rows.Err(); err != nil {
		{{ if (hasID) }} return nil, errors.Wrap(err, "rows iteration error") {{ else }} return errors.Wrap(err, "rows iteration error") {{ end }}
	}

	return {{ if (hasID) }} returnIDs, nil {{ else }} nil {{ end }}
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
		return nil, errors.Wrap(err, "failed to get value of {{ $field | fieldName }}")
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
		{{ if (hasID) }} return nil, errors.Wrap(err, "failed to build query") {{ else }} return errors.Wrap(err, "failed to build query") {{ end }}
	}
	t.logQuery(ctx, sqlQuery, args)

	{{ if (hasID) }}var id {{IDType}}
	err = t.DB(ctx, true).QueryRowContext(ctx,sqlQuery, args...).Scan(&id) {{ else }} _, err = t.DB(ctx, true).ExecContext(ctx,sqlQuery, args...) {{ end }}
	if err != nil {
		if IsPgUniqueViolation(err) {
			{{ if (hasID) }}return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ else }}return errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error()) {{ end }}
		}

		{{ if (hasID) }} return nil, errors.Wrap(err, "failed to create {{ structureName }}") {{ else }} return errors.Wrap(err, "failed to create {{ structureName }}") {{ end }}
	}

	{{ if (hasID) }}
	{{- range $index, $field := fields }}
	{{- if and ($field | isRelation) ($field | relationAllowSubCreating) }}
	    if options.relations && model.{{ $field | fieldName }} != nil { {{ if ($field | isRepeated) }}
			for _, item := range model.{{ $field | fieldName }} {
				item.{{ $field | getRefID }} = id
				s, err := New{{ $field | relationStorageName }}(t.config)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create {{ $field | fieldName }}")
				}

                {{ if ($field | hasIDFromRelation) }} _, err = s.Create(ctx, item) {{ else }} err = s.Create(ctx, item) {{ end }}
				if err != nil {
				{{ if (hasID) }} return nil, errors.Wrap(err, "failed to create {{ $field | fieldName }}") {{ else }} return errors.Wrap(err, "failed to create {{ $field | fieldName }}") {{ end }}
				}
			} {{ else }}
			s, err := New{{ $field | relationStorageName }}(t.config)
			if err != nil {
				{{ if (hasID) }} return nil, errors.Wrap(err, "failed to create {{ $field | fieldName }}") {{ else }} return errors.Wrap(err, "failed to create {{ $field | fieldName }}") {{ end }}
			}

			model.{{ $field | fieldName }}.{{ $field | getRefID }} = id
			{{ if ($field | hasIDFromRelation) }} _, err = s.Create(ctx, model.{{ $field | fieldName }}) {{ else }} err = s.Create(ctx, model.{{ $field | fieldName }}) {{ end }}
			if err != nil {
				{{ if (hasID) }} return nil, errors.Wrap(err, "failed to create {{ $field | fieldName }}") {{ else }} return errors.Wrap(err, "failed to create {{ $field | fieldName }}") {{ end }}
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
	{{- if (hasID) }}
	Create(ctx context.Context, model *{{structureName}}, opts ...Option) (*{{IDType}}, error)
	{{- else }} 
	Create(ctx context.Context, model *{{structureName}}, opts ...Option) error
	{{- end }}
	{{ if (hasID) }}BatchCreate(ctx context.Context, models []*{{structureName}}, opts ...Option) ([]string, error)
	{{- else }}
	BatchCreate(ctx context.Context, models []*{{structureName}}, opts ...Option) error
	{{- end }}
	Update(ctx context.Context, id {{IDType}}, updateData *{{structureName}}Update) error
	{{- if (hasPrimaryKey) }}
	DeleteBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, {{getPrimaryKey.GetName}} {{IDType}}, opts ...Option) error
	{{- end }}
	{{- if (hasPrimaryKey) }}
	FindBy{{ getPrimaryKey.GetName | camelCase }}(ctx context.Context, id {{IDType}}, opts ...Option) (*{{ structureName }}, error)
	{{- end }}
}

// {{structureName}}SearchOperations is an interface for searching the {{ tableName }} table.
type {{structureName}}SearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*{{structureName}}, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*{{structureName}}, error)
}

// {{structureName}}PaginationOperations is an interface for pagination operations.
type {{structureName}}PaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*{{structureName}}, *Paginator, error)
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

// {{structureName}}AdvancedDeletion is an interface for advanced deletion operations.
type {{structureName}}AdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// {{structureName}}RawQueryOperations is an interface for executing raw queries.
type {{structureName}}RawQueryOperations interface {
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
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
	{{structureName}}AdvancedDeletion
	{{structureName}}RawQueryOperations
}

// New{{ storageName }} returns a new {{ storageName | lowerCamelCase }}.
func New{{ storageName }}(config *Config) ({{ storageName }}, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.DB == nil {
		return nil, errors.New("config.DB is nil")
	}
	if config.DB.DBRead == nil {
		return nil, errors.New("config.DB.DBRead is nil")
	}
	if config.DB.DBWrite == nil {
		config.DB.DBWrite = config.DB.DBRead
	}

	return &{{ storageName | lowerCamelCase }}{
		config: config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
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
func (t *{{ storageName | lowerCamelCase }}) DB(ctx context.Context, isWrite bool) QueryExecer {
	var db QueryExecer

	// Check if there is an active transaction in the context.
	if tx, ok := TxFromContext(ctx); ok {
		if tx == nil {
			t.logError(ctx, errors.New("transaction is nil"), "failed to get transaction from context")
			// set default connection
			return t.config.DB.DBWrite
		}

		return tx
	}

	// Use the appropriate connection based on the operation type.
	if isWrite {
		db = t.config.DB.DBWrite
	} else {
		db = t.config.DB.DBRead
	}

	return db
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

	_, err := t.DB(ctx, true).ExecContext(ctx,sqlQuery)
	return err
}

// DropTable drops the table.
func (t *{{ storageName | lowerCamelCase }}) DropTable(ctx context.Context) error {
	sqlQuery := ` + "`" + `
		DROP TABLE IF EXISTS {{ tableName }};
	` + "`" + `

	_, err := t.DB(ctx, true).ExecContext(ctx,sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *{{ storageName | lowerCamelCase }}) TruncateTable(ctx context.Context) error {
	sqlQuery := ` + "`" + `
		TRUNCATE TABLE {{ tableName }};
	` + "`" + `

	_, err := t.DB(ctx, true).ExecContext(ctx,sqlQuery)
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
