package templater

import (
	"fmt"
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
	"strings"
	"text/template"

	importpkg "github.com/cjp2600/structify/plugin/import"
	helperpkg "github.com/cjp2600/structify/plugin/pkg/helper"
	tmplpkg "github.com/cjp2600/structify/plugin/provider/postgres/tmpl"
	statepkg "github.com/cjp2600/structify/plugin/state"
)

// tableTemplater is the templater for the init statement.
// It implements the state.Templater interface.
type tableTemplater struct {
	state   *statepkg.State
	message *descriptorpb.DescriptorProto
}

// NewTableTemplater returns a new initTemplater.
func NewTableTemplater(message *descriptorpb.DescriptorProto, state *statepkg.State) statepkg.Templater {
	return &tableTemplater{
		state:   state,
		message: message,
	}
}

// BuildTemplate builds the template.
func (t *tableTemplater) BuildTemplate() string {
	tmpl, err := helperpkg.ExecuteTemplate(
		tmplpkg.TableTemplate,
		t.Funcs(),
		t,
		helperpkg.IncludeTemplate{
			Name: "storage",
			Body: tmplpkg.TableStorageTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "structure",
			Body: tmplpkg.StructureTemplate,
		},
	)
	if err != nil {
		log.Fatalf("failed to execute template: %v", err)
		return ""
	}

	return tmpl
}

// Imports returns the imports.
func (t *tableTemplater) Imports() importpkg.ImportSet {
	is := importpkg.ImportSet{}
	is.Enable(
		importpkg.ImportDb,
		importpkg.ImportLibPQ,
		importpkg.ImportStrconv,
		importpkg.ImportFMT,
		importpkg.ImportErrors,
		importpkg.ImportSquirrel,
	)

	return is
}

// Funcs returns the template functions.
func (t *tableTemplater) Funcs() map[string]interface{} {
	return template.FuncMap{

		// fieldType returns the field type.
		"fieldType": func(f *descriptorpb.FieldDescriptorProto) string {
			// if the field is a single type, return the single type.
			if t.state.SingleTypes.ExistByName(f.GetName()) {
				mds := t.state.SingleTypes.GetByName(f.GetName())
				if mds != nil {
					fieldType := mds.FieldType
					if helperpkg.IsOptional(f) {
						if !strings.Contains(mds.FieldType, "*") {
							fieldType = fmt.Sprintf("*%s", fieldType)
						}
					}
					return fieldType
				}
			}

			// if the field is a nested message, return the structure name.
			if t.state.NestedMessages.IsJSON(f) {
				md := t.state.NestedMessages.GetByFieldDescriptor(f)
				if md != nil {
					return helperpkg.TypePrefix(f, md.StructureName)
				}
			}

			return helperpkg.ConvertType(f)
		},

		// comment returns the comment.
		"comment": func() string {
			if opts := helperpkg.GetMessageOptions(t.message); opts != nil {
				if opts.Comment != "" {
					return opts.Comment
				}
			}
			return ""
		},

		// isLastField returns true if the field is the last field.
		"isLastField": func(f *descriptorpb.FieldDescriptorProto) bool {
			return f == t.message.Field[len(t.message.Field)-1]
		},

		// getDefaultValue returns the default value.
		"getDefaultValue": func(f *descriptorpb.FieldDescriptorProto) string {
			if opts := helperpkg.GetFieldOptions(f); opts != nil {
				return opts.GetDefault()
			}
			return ""
		},

		// isPrimaryKey returns true if the field is primary key.
		"isPrimaryKey": func(f *descriptorpb.FieldDescriptorProto) bool {
			if opts := helperpkg.GetFieldOptions(f); opts != nil {
				return opts.GetPrimaryKey()
			}
			return false
		},

		// isUnique returns true if the field is unique.
		"isUnique": func(f *descriptorpb.FieldDescriptorProto) bool {
			if opts := helperpkg.GetFieldOptions(f); opts != nil {
				return opts.GetUnique()
			}
			return false
		},

		// isNotNull returns true if the field is not null.
		"isNotNull": func(f *descriptorpb.FieldDescriptorProto) bool {
			if opts := helperpkg.GetFieldOptions(f); opts != nil {
				return opts.GetNullable() == false
			}
			return false
		},

		// isAutoIncrement returns true if the field is auto increment.
		"isAutoIncrement": func(f *descriptorpb.FieldDescriptorProto) bool {
			if opts := helperpkg.GetFieldOptions(f); opts != nil {
				return opts.GetAutoIncrement()
			}
			return false
		},

		// postgresType returns the postgres type.
		"postgresType": func(f *descriptorpb.FieldDescriptorProto) string {
			return helperpkg.PostgresType(helperpkg.ConvertType(f), helperpkg.GetFieldOptions(f), t.state.NestedMessages.IsJSON(f))
		},

		// storageName returns the upper camel case storage name.
		"storageName": func() string {
			return fmt.Sprintf("%sStorage", helperpkg.UpperCamelCase(t.message.GetName()))
		},

		// fields returns the fields.
		"fields": func() []*descriptorpb.FieldDescriptorProto {
			return t.message.GetField()
		},

		"fieldsByMessage": func(message *descriptorpb.DescriptorProto) []*descriptorpb.FieldDescriptorProto {
			return message.GetField()
		},

		// fieldName returns the upper camel case field name.
		"fieldName": func(f *descriptorpb.FieldDescriptorProto) string {
			return helperpkg.UpperCamelCase(f.GetName())
		},

		// isRelation returns the field type.
		"isRelation": func(f *descriptorpb.FieldDescriptorProto) bool {
			return t.state.NestedMessages.CheckIsRelation(f)
		},

		// sourceName returns the source name.
		"sourceName": func(f *descriptorpb.FieldDescriptorProto) string {
			return f.GetName()
		},

		// structureName returns the upper camel case structure name.
		"structureName": func() string {
			return helperpkg.UpperCamelCase(t.message.GetName())
		},

		// tableName returns the table name.
		"tableName": func() string {
			if opts := helperpkg.GetMessageOptions(t.message); opts != nil {
				if opts.Table != "" {
					return opts.Table
				}
			}
			return helperpkg.Plural(t.message.GetName())
		},

		// tableComment returns the table comment.
		"tableComment": func() string {
			if opts := helperpkg.GetMessageOptions(t.message); opts != nil {
				if opts.Comment != "" {
					return opts.Comment
				}
			}
			return ""
		},

		// clientName returns the upper camel case client name.
		"clientName": func() string {
			return fmt.Sprintf("%s%s", helperpkg.UpperCamelCase(t.state.FileName), DBClientPostfix)
		},

		// clientName_lower returns the lower camel case client name.
		"clientName_lower": func() string {
			return fmt.Sprintf("%s%s", helperpkg.LowerCamelCase(t.state.FileName), DBClientPostfix)
		},

		// camelCase returns the upper camel case.
		"camelCase": helperpkg.UpperCamelCase,

		// lowerCamelCase returns the lower camel case.
		"lowerCamelCase": helperpkg.LowerCamelCase,
	}
}
