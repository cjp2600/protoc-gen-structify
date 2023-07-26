package plugin

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

// createNewPostgresTableTemplate creates a new PostgresTable from the given descriptor.
// It also updates the state with the new table.
func createNewPostgresTableTemplate(d *descriptorpb.DescriptorProto, state *State) Templater {
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
		structureName := detectStructName(convertedType)
		relKey := d.GetName() + "::" + structureName

		// detect id field
		if strings.ToLower(f.GetName()) == "id" || (options != nil && options.PrimaryKey) {
			if options != nil && !options.PrimaryKey {
				options.PrimaryKey = true
			}

			// remove relation if exist
			table.IdType = convertedType
			table.IdName = sourceName
			table.IsIdUUID = postgresType(convertedType, options) == "UUID"
		}

		// detect relation field and change type
		field := &Field{
			Name:       sToCml(f.GetName()),
			SourceName: sourceName,
			Type:       convertedType,
			DBType:     postgresType(convertedType, options),
			IsRelation: checkIsRelation(f),
			Optional:   isOptional(f),
		}

		field.Options = Options{}
		// detect relation field and change type
		if checkIsRelation(f) {
			table.HasRelations = true
			if v, ok := state.Relations[relKey]; ok {
				field.Options.Relation = v
			}
		}

		// detect optional field
		if isOptional(f) {
			field.Options.Nullable = true // set nullable option
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
			field.Type = buildJSONTypeName(d.GetName(), structureName)
			if field.Optional {
				field.Type = "*" + field.Type
			}

			field.Options.JSON = true // set json option
			field.DBType = "JSONB"    // set jsonb type
		}

		// change type if json type
		if v, ok := state.JSONTypes[sToCml(d.GetName())+"::"+field.SourceName]; ok {
			field.Type = "*" + v.TypeName
		}

		// detect repeated objects field and change type
		if isRepeatedObjectField(f, *field, state) {
			field.Type = field.Type + "Repeated"
		}

		// detect uuid field and change type
		if options != nil {
			field.Options.PrimaryKey = options.PrimaryKey
			field.Options.Unique = options.Unique
			field.Options.Nullable = options.Nullable
			field.Options.AutoIncrement = options.AutoIncrement
			field.Options.Default = options.Default
			field.Options.UUID = options.Uuid
			field.Options.JSON = options.Json
		}

		// add to table
		table.Fields = append(table.Fields, field)
	}

	var columns []string
	for _, field := range table.Fields {
		if !field.IsRelation {
			columns = append(columns, lowerCasePlural(d.GetName())+"."+field.SourceName)
		}
	}
	table.Columns = columns

	return table
}

type PostgresTable struct {
	Name         string
	TableName    string
	IdType       string
	IdName       string
	IsIdUUID     bool
	Fields       []*Field
	Columns      []string
	CreateSQL    string
	Comment      string
	HasRelations bool
}

// Imports - returns the imports for the template.
func (t *PostgresTable) Imports() ImportSet {
	return ImportSet{
		ImportLibPQ:   true,
		ImportDb:      true,
		ImportStrings: false,
		ImportSync:    true,
	}
}

func (t *PostgresTable) BuildTemplate() string {
	t.CreateSQL = t.GenerateCreateSQL()

	if t.IsIdUUID {
		// Add the uuid extension to the create sql.
		t.CreateSQL = ExtensionUUID.String() + "\n" + t.CreateSQL
	}

	var output bytes.Buffer

	funcs := template.FuncMap{
		"firstLetter":     firstLetterLower,
		"sliceToString":   sliceToString,
		"sToLowerCamel":   sToLowerCamel,
		"sToCml":          sToCml,
		"lowerCasePlural": lowerCasePlural,
		"lowerCase":       lowerCase,
		"isContainsStar":  isContainsStar,
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
	IsRelation bool
}

type Options struct {
	PrimaryKey    bool
	Unique        bool
	Nullable      bool
	AutoIncrement bool
	Default       string
	UUID          bool
	JSON          bool
	Relation      *Relation
}

const createSQLTemplate = `CREATE TABLE IF NOT EXISTS {{.TableName}} (
{{- range $index, $element := .Fields}}
{{- if not $element.IsRelation}}
{{$element.SourceName}} {{if $element.Options.AutoIncrement}}{{$element.DBType}} SERIAL{{else}}{{$element.DBType}}{{end}}{{if $element.Options.PrimaryKey}} PRIMARY KEY{{end}}{{if $element.Options.Unique}} UNIQUE{{end}}{{if not $element.Options.Nullable}} NOT NULL{{end}}{{if $element.Options.Default}} DEFAULT {{$element.Options.Default}}{{end}}{{if not (isLast $index (len $.Fields))}},{{end}}
{{- end}}
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
