package plugin

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

// createNewPostgresTableTemplate creates a new PostgresTable from the given descriptor.
func createNewPostgresTableTemplate(d *descriptorpb.DescriptorProto) Templater {
	table := &PostgresTable{
		Name: d.GetName(),
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
		field := &Field{
			Name:       sToCml(f.GetName()),
			SourceName: f.GetName(),
			Type:       convertType(f),
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
`

type PostgresTable struct {
	Name      string
	TableName string
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
}

const createSQLTemplate = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
{{range $index, $element := .Fields}}	{{$element.SourceName}} {{goTypeToPostgresType $element.Type}}{{if not (isLast $index (len $.Fields))}},{{end}}
{{end}});{{if .Comment}}COMMENT ON TABLE {{.TableName}} IS '{{.Comment }}';{{end}}
`

// GenerateCreateSQL generates the SQL statement to create the table.
// This is used in the CreateTableSQL method. It is also used in the template
func (t *PostgresTable) GenerateCreateSQL() string {
	var output bytes.Buffer

	funcs := template.FuncMap{
		"isLast": func(x, total int) bool {
			return x == (total - 1)
		},
		"goTypeToPostgresType": goTypeToPostgresType,
	}

	tmpl, err := template.New("createSQL").Funcs(funcs).Parse(createSQLTemplate)
	if err != nil {
		return ""
	}

	if err = tmpl.Execute(&output, t); err != nil {
		return ""
	}

	return output.String()
}
