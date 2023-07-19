package plugin

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

// createNewPostgresTableTemplate creates a new PostgresTable from the given descriptor.
func createNewPostgresTableTemplate(d *descriptorpb.DescriptorProto) Templater {
	table := &PostgresTable{
		Name:      d.GetName(),
		TableName: lowerCasePlural(d.GetName()),
	}

	if opts := getMessageOptions(d); opts != nil {
		if opts.Table != "" {
			table.TableName = opts.Table
		}
		if opts.Comment != "" {
			table.Comment = opts.Comment
		}
	}

	for _, f := range d.GetField() {
		options := getFieldOptions(f)

		convertedType := convertType(f)
		sourceName := f.GetName()

		// detect id field
		if strings.ToLower(f.GetName()) == "id" || (options != nil && options.PrimaryKey) {
			if options != nil && !options.PrimaryKey {
				options.PrimaryKey = true
			}

			table.IdType = convertedType
			table.IdName = sourceName
		}

		field := &Field{
			Name:       sToCml(f.GetName()),
			SourceName: sourceName,
			Type:       convertedType,
			DBType:     postgresType(convertedType, options),
		}

		if options != nil {
			field.Options = Options{
				PrimaryKey:    options.PrimaryKey,
				Unique:        options.Unique,
				Nullable:      options.Nullable,
				AutoIncrement: options.AutoIncrement,
				Default:       options.Default,
			}
		}

		table.Fields = append(table.Fields, field)
	}

	table.Columns = make([]string, len(table.Fields))
	for i, field := range table.Fields {
		table.Columns[i] = field.SourceName
	}

	return table
}

// PostgresStructTemplate is the template for the Go struct.
const PostgresStructTemplate = `
type {{.Name | sToLowerCamel }}Store struct {
	db *sql.DB
}

// {{.Name}} is a struct for the "{{.TableName}}" table.
type {{.Name}} struct {
{{range .Fields}}
    {{.Name}} {{.Type}}` + " `db:\"{{.SourceName}}\"`" + `{{end}}
}

// TableName returns the name of the table.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) TableName() string {
	return "{{.TableName}}"
}

// Columns returns the database columns for the table.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) Columns() []string {
	return {{.Columns | sliceToString}}
}

// CreateTableSQL returns the SQL statement to create the table.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) CreateTableSQL() string {
	return ` + "`" + `{{.CreateSQL}}` + "`" + `
}

{{if .IdType}}
// FindByID returns a single row by ID.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) FindBy{{ .IdName  | sToCml }}({{ .IdName }} {{ .IdType }}) (*{{.Name | sToCml }}, error) {
	return {{.Name | firstLetter}}.FindOne({{.Name | sToCml }}{{ .IdName  | sToCml }}Eq({{ .IdName }}))
}
{{end}}

// FindOne filters rows by the provided conditions and returns the first matching row.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) FindOne(conditions ...Condition) (*{{.Name}}, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Select({{.Name | firstLetter}}.Columns()...).From({{.Name | firstLetter}}.TableName())

	for _, condition := range conditions {
		query = condition.Apply(query)
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql: %w", err)
	}

	row := {{.Name | firstLetter}}.db.QueryRow(sqlQuery, args...)

	var model {{.Name}}
	err = row.Scan({{range .Fields}}&model.{{.Name}}, {{end}})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &model, nil
}

// FindMany filters rows by the provided conditions and returns matching rows.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) FindMany(conditions ...Condition) ([]{{.Name}}, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Select({{.Name | firstLetter}}.Columns()...).From({{.Name | firstLetter}}.TableName())

	for _, condition := range conditions {
		query = condition.Apply(query)
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql: %w", err)
	}

	rows, err := {{.Name | firstLetter}}.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var {{.Name | lowerCasePlural}} []{{.Name}}
	for rows.Next() {
		var model {{.Name}}
		err = rows.Scan({{range .Fields}}&model.{{.Name}}, {{end}})
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		{{.Name | lowerCasePlural}} = append({{.Name | lowerCasePlural}}, model)
	}

	return {{.Name | lowerCasePlural}}, nil
}

// Count returns the number of rows that match the provided conditions.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) Count(conditions ...Condition) (int64, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Select("COUNT(*)").From({{.Name | firstLetter}}.TableName())

	for _, condition := range conditions {
		query = condition.Apply(query)
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql: %w", err)
	}

	row := {{.Name | firstLetter}}.db.QueryRow(sqlQuery, args...)

	var count int64
	err = row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	return count, nil
}

// Delete deletes rows that match the provided conditions.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) Delete(conditions ...Condition) (int64, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Delete({{.Name | firstLetter}}.TableName())

	for _, condition := range conditions {
		query = condition.ApplyDelete(query)
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql: %w", err)
	}

	result, err := {{.Name | firstLetter}}.db.Exec(sqlQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	deletedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return deletedRows, nil
}

// DeleteWithTx deletes rows that match the provided conditions inside a transaction.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel }}Store) DeleteWithTx(tx *sql.Tx, conditions ...Condition) (int64, error) {
	if tx == nil {
		return 0, ErrNoTransaction
	}
	
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Delete({{.Name | firstLetter}}.TableName())

	for _, condition := range conditions {
		query = condition.ApplyDelete(query)
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql: %w", err)
	}

	result, err := tx.Exec(sqlQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	deletedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return deletedRows, nil
}

{{if .IdType}}
// {{.Name}}UpdateRequest is the data required to update a row.
type {{.Name}}UpdateRequest struct {
{{range .Fields}}{{if not (eq (.Name | sToLowerCamel) ($.IdName | sToLowerCamel) )}}
	{{.Name}} *{{.Type}}{{end}}{{end}}
}

// Update updates a row with the provided data.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel}}Store) Update(ctx context.Context, id {{.IdType}}, model *{{.Name}}UpdateRequest) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Update({{.Name | firstLetter}}.TableName())

	{{range .Fields}}
	{{if not .Options.PrimaryKey}}
	if model.{{.Name}} != nil {
		query = query.Set("{{.SourceName}}", model.{{.Name}})
	}
	{{end}}
	{{end}}

	query = query.Where(sq.Eq{"{{.IdName}}": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql: %w", err)
	}

	_, err = {{.Name | firstLetter}}.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// UpdateWithTx updates a row with the provided data inside a transaction.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel}}Store) UpdateWithTx(ctx context.Context, tx *sql.Tx, id {{.IdType}}, model *{{.Name}}UpdateRequest) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Update({{.Name | firstLetter}}.TableName())

	{{range .Fields}}
	{{if not .Options.PrimaryKey}}
	if model.{{.Name}} != nil {
		query = query.Set("{{.SourceName}}", model.{{.Name}})
	}
	{{end}}
	{{end}}

	query = query.Where(sq.Eq{"{{$.IdName}}": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build sql: %w", err)
	}

	_, err = tx.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// Create inserts a new row into the database.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel}}Store) Create(ctx context.Context, model *{{.Name}}) ({{.IdType}}, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Insert({{.Name | firstLetter}}.TableName()).
		Columns({{range .Fields}}{{if not .Options.PrimaryKey}}"{{.SourceName}}", {{end}}{{end}}).
		Suffix("RETURNING \"{{.IdName}}\"").
		Values({{range .Fields}}{{if not .Options.PrimaryKey}}model.{{.Name}}, {{end}}{{end}})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return {{if .IdType}}""{{end}}, fmt.Errorf("failed to build sql: %w", err)
	}

	row := {{.Name | firstLetter}}.db.QueryRowContext(ctx, sqlQuery, args...)
	if err != nil {
		return {{if .IdType}}""{{end}}, fmt.Errorf("failed to execute query: %w", err)
	}

	var id {{.IdType}}
	if err := row.Scan(&id); err != nil {
		return {{if .IdType}}""{{end}}, fmt.Errorf("failed to scan id: %w", err)
	}

	return id, nil
}

// CreateWithTx inserts a new row into the database inside a transaction.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel}}Store) CreateWithTx(tx *sql.Tx, model *{{.Name}}) ({{.IdType}}, error) {
	if tx == nil {
		return {{if .IdType}}""{{end}}, ErrNoTransaction
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Insert({{.Name | firstLetter}}.TableName()).
		Columns({{range .Fields}}{{if not .Options.PrimaryKey}}"{{.SourceName}}", {{end}}{{end}}).
		Suffix("RETURNING \"{{.IdName}}\"").
		Values({{range .Fields}}{{if not .Options.PrimaryKey}}model.{{.Name}}, {{end}}{{end}})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return {{if .IdType}}""{{end}}, fmt.Errorf("failed to build sql: %w", err)
	}

	row := tx.QueryRow(sqlQuery, args...)
	if err != nil {
		return {{if .IdType}}""{{end}}, fmt.Errorf("failed to execute query: %w", err)
	}

	var id {{.IdType}}
	if err := row.Scan(&id); err != nil {
		return {{if .IdType}}""{{end}}, fmt.Errorf("failed to scan id: %w", err)
	}

	return id, nil
}

// CreateMany inserts multiple rows into the database.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel}}Store) CreateMany(ctx context.Context, models []*{{.Name}}) ([]{{.IdType}}, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Insert({{.Name | firstLetter}}.TableName()).
		Columns({{range .Fields}}{{if not .Options.PrimaryKey}}"{{.SourceName}}", {{end}}{{end}})

	for _, model := range models {
		query = query.Values({{range .Fields}}{{if not .Options.PrimaryKey}}model.{{.Name}}, {{end}}{{end}})
	}

	query = query.Suffix("RETURNING \"{{.IdName}}\"")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql: %w", err)
	}

	rows, err := {{.Name | firstLetter}}.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var ids []{{.IdType}}
	for rows.Next() {
		var id {{.IdType}}
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return ids, nil
}

// CreateManyWithTx inserts multiple rows into the database inside a transaction.
func ({{.Name | firstLetter}} *{{.Name | sToLowerCamel}}Store) CreateManyWithTx(ctx context.Context, tx *sql.Tx, models []*{{.Name}}) ([]{{.IdType}}, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Insert({{.Name | firstLetter}}.TableName()).
		Columns({{range .Fields}}{{if not .Options.PrimaryKey}}"{{.SourceName}}", {{end}}{{end}})

	for _, model := range models {
		query = query.Values({{range .Fields}}{{if not .Options.PrimaryKey}}model.{{.Name}}, {{end}}{{end}})
	}

	query = query.Suffix("RETURNING \"{{.IdName}}\"")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql: %w", err)
	}

	rows, err := tx.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var ids []{{.IdType}}
	for rows.Next() {
		var id {{.IdType}}
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return ids, nil
}
{{end}}
`

type PostgresTable struct {
	Name      string
	TableName string
	IdType    string
	IdName    string
	Fields    []*Field
	Columns   []string
	CreateSQL string
	Comment   string
}

// Imports - returns the imports for the template.
func (t *PostgresTable) Imports() ImportSet {
	return ImportSet{
		ImportLibPQ:   true,
		ImportDb:      true,
		ImportStrings: false,
	}
}

func (t *PostgresTable) BuildTemplate() string {
	t.CreateSQL = t.GenerateCreateSQL()

	var output bytes.Buffer

	funcs := template.FuncMap{
		"firstLetter":     firstLetterLower,
		"sliceToString":   sliceToString,
		"sToLowerCamel":   sToLowerCamel,
		"sToCml":          sToCml,
		"lowerCasePlural": lowerCasePlural,
		"lowerCase":       lowerCase,
	}

	tmpl, err := template.New("goFile").Funcs(funcs).Parse(PostgresStructTemplate)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if err = tmpl.Execute(&output, t); err != nil {
		fmt.Println(err)
		return ""
	}

	return output.String()
}

type Field struct {
	Name       string
	SourceName string
	Type       string
	DBType     string
	Optional   bool
	Options    Options
}

type Options struct {
	PrimaryKey    bool
	Unique        bool
	Nullable      bool
	AutoIncrement bool
	Default       string
}

const createSQLTemplate = `CREATE TABLE IF NOT EXISTS {{.TableName}} (
{{- range $index, $element := .Fields}}
{{$element.SourceName}} {{if $element.Options.AutoIncrement}}{{$element.DBType}} SERIAL{{else}}{{$element.DBType}}{{end}}{{if $element.Options.PrimaryKey}} PRIMARY KEY{{end}}{{if $element.Options.Unique}} UNIQUE{{end}}{{if not $element.Options.Nullable}} NOT NULL{{end}}{{if $element.Options.Default}} DEFAULT {{$element.Options.Default}}{{end}}{{if not (isLast $index (len $.Fields))}},{{end}}
{{- end}});{{if .Comment}}COMMENT ON TABLE {{.TableName}} IS '{{.Comment}}';{{end}}`

// GenerateCreateSQL generates the SQL statement to create the table.
// This is used in the CreateTableSQL method. It is also used in the template
func (t *PostgresTable) GenerateCreateSQL() string {
	var output bytes.Buffer

	funcs := template.FuncMap{
		"isLast": func(x, total int) bool {
			return x == (total - 1)
		},
	}

	tmpl, err := template.New("createSQL").Funcs(funcs).Parse(createSQLTemplate)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if err = tmpl.Execute(&output, t); err != nil {
		fmt.Println(err)
		return ""
	}

	return output.String()
}
