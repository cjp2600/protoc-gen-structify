package plugin

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

type NestedTable struct {
	Name   string
	Fields []*Field
}

const NestedTableStructTemplate = `
// {{.Name}} is a nested table.
type {{.Name}} struct {
{{range .Fields}}
	{{.Name}} {{.Type}}` + " `json:\"{{.SourceName}}\"`" + `{{end}}
}

// Scan implements the sql.Scanner interface for MyJSONType
func (m *{{.Name}}) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return fmt.Errorf("can't convert %T to {{.Name}}", src)
}

// Value implements the driver.Valuer interface for {{.Name}}
func (m *{{.Name}}) Value() (driver.Value, error) {
	return json.Marshal(m)
}`

func (n NestedTable) BuildTemplate() string {
	var output bytes.Buffer

	funcs := template.FuncMap{
		"firstLetter":     firstLetterLower,
		"sliceToString":   sliceToString,
		"sToLowerCamel":   sToLowerCamel,
		"sToCml":          sToCml,
		"lowerCasePlural": lowerCasePlural,
		"lowerCase":       lowerCase,
	}

	tmpl, err := template.New("goFile").Funcs(funcs).Parse(NestedTableStructTemplate)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if err = tmpl.Execute(&output, n); err != nil {
		fmt.Println(err)
		return ""
	}

	return output.String()
}

func (n NestedTable) Imports() ImportSet {
	return ImportSet{}
}

// createNewNestedTemplate creates a new nested template.
// It returns a new nested template.
func createNewNestedTemplate(d *descriptorpb.DescriptorProto, state *State) Templater {
	table := &NestedTable{
		Name: d.GetName(),
	}

	for _, f := range d.GetField() {
		options := getFieldOptions(f)
		convertedType := convertType(f)
		sourceName := f.GetName()
		structureName := detectStructName(convertedType)
		relKey := d.GetName() + "::" + structureName

		field := &Field{
			Name:       sToCml(f.GetName()),
			SourceName: sourceName,
			Type:       convertedType,
			DBType:     postgresType(convertedType, options),
			IsRelation: checkIsRelation(f),
			Optional:   isOptional(f),
		}

		// detect json field and change type
		if isJSON(f, state) {

			// remove relation
			delete(state.Relations, relKey)
			{
				field.IsRelation = false
				field.Options.Relation = nil
			}

			// change type
			field.Type = d.GetName() + sToCml(structureName)
			field.Options.JSON = true // set json option
		}

		table.Fields = append(table.Fields, field)
	}

	return table
}
