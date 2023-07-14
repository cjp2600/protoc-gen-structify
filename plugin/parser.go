package plugin

import (
	"bytes"
	"google.golang.org/protobuf/types/descriptorpb"
	"html/template"
)

const structTemplate = `
type {{.Name}} struct {
    {{range .Fields}}
        {{.Name}} {{.Type}}
    {{end}}
}
`

type Table struct {
	Name      string
	TableName string
	Fields    []*Field
}

func (t *Table) String() string {
	var output bytes.Buffer

	tmpl, err := template.New("goFile").Parse(structTemplate)
	if err != nil {
		return ""
	}

	err = tmpl.Execute(&output, t)
	if err != nil {
		return ""
	}

	return output.String()
}

type Field struct {
	Name string
	Type string
}

// ParseProto обрабатывает .proto файлы из запроса и преобразует их в структуры Table.
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

// newTable создает новую структуру Table из DescriptorProto.
func newTable(d *descriptorpb.DescriptorProto) *Table {
	table := &Table{
		Name: d.GetName(),
	}

	// Извлекаем опции сообщения, если они есть.
	if opts := getMessageOptions(d); opts != nil {
		if opts.Table != "" {
			table.TableName = opts.Table
		}
	}

	for _, f := range d.GetField() {
		field := &Field{
			Name: sToCml(f.GetName()),
			Type: prepareType(f.GetType().String()),
		}

		table.Fields = append(table.Fields, field)
	}

	return table
}
