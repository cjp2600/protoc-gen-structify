package plugin

import (
	"bytes"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

// newPostgresTable creates a new PostgresTable from the given descriptor.
func newPostgresTable(d *descriptorpb.DescriptorProto) Templater {
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
// {{.Name}} is a struct for the "{{.TableName}}" table.
type {{.Name}} struct {
{{range .Fields}}
    {{.Name}} {{.Type}}` + " `db:\"{{.SourceName}}\"`" + `{{end}}
}

// TableName returns the name of the table.
func ({{.Name | firstLetter}} *{{.Name}}) TableName() string {
	return "{{.TableName}}"
}

// Columns returns the database columns for the table.
func ({{.Name | firstLetter}} *{{.Name}}) Columns() []string {
	return {{.Columns | sliceToString}}
}

// CreateTableSQL returns the SQL statement to create the table.
func ({{.Name | firstLetter}} *{{.Name}}) CreateTableSQL() string {
	return ` + "`" + `{{.CreateSQL}}` + "`" + `
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

func (t *PostgresTable) BuildTemplate() string {
	t.CreateSQL = t.GenerateCreateSQL()

	var output bytes.Buffer

	funcs := template.FuncMap{
		"firstLetter":   firstLetterLower,
		"sliceToString": sliceToString,
	}

	tmpl, err := template.New("goFile").Funcs(funcs).Parse(PostgresStructTemplate)
	if err != nil {
		return ""
	}

	if err = tmpl.Execute(&output, t); err != nil {
		return ""
	}

	return output.String()
}

type Field struct {
	Name       string
	SourceName string
	Type       string
}
