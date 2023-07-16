package plugin

import (
	"bytes"
	"google.golang.org/protobuf/types/descriptorpb"
	"html/template"
)

// ParseProto parses the proto file and returns a slice of Tables.
func (p *Plugin) ParseProto() ([]*Table, error) {
	var tables []*Table

	for _, f := range p.req.GetProtoFile() {
		for _, m := range f.GetMessageType() {
			if !isUserMessage(f, m) {
				continue
			}

			table := newTable(m)
			tables = append(tables, table)
		}
	}

	return tables, nil
}

// newTable creates a new Table from the given descriptor.
func newTable(d *descriptorpb.DescriptorProto) *Table {
	table := &Table{
		Name: d.GetName(),
	}

	if opts := getMessageOptions(d); opts != nil {
		if opts.Table != "" {
			table.TableName = opts.Table
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

// structTemplate is the template for the Go struct.
const structTemplate = `
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

type Table struct {
	Name      string
	TableName string
	Fields    []*Field
	Columns   []string
	CreateSQL string
	Comment   string
}

func (t *Table) String() string {
	t.CreateSQL = t.GenerateCreateSQL()

	var output bytes.Buffer

	funcs := template.FuncMap{
		"firstLetter":   firstLetterLower,
		"sliceToString": sliceToString,
	}

	tmpl, err := template.New("goFile").Funcs(funcs).Parse(structTemplate)
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
