package plugin

import (
	"bytes"
	"html/template"
)

const createSQLTemplate = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
{{range $index, $element := .Fields}}	{{$element.SourceName}} {{goTypeToPostgresType $element.Type}}{{if not (isLast $index (len $.Fields))}},{{end}}
{{end}});
{{if .Comment}}; COMMENT ON TABLE {{.TableName}} IS '{{.Comment}}'{{end}}
`

// GenerateCreateSQL generates the SQL statement to create the table.
// This is used in the CreateTableSQL method. It is also used in the template
func (t *Table) GenerateCreateSQL() string {
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
