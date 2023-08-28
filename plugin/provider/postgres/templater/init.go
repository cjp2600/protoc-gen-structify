package templater

import (
	"fmt"
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
	"text/template"

	importpkg "github.com/cjp2600/protoc-gen-structify/plugin/import"
	helperpkg "github.com/cjp2600/protoc-gen-structify/plugin/pkg/helper"
	tmplpkg "github.com/cjp2600/protoc-gen-structify/plugin/provider/postgres/tmpl"
	statepkg "github.com/cjp2600/protoc-gen-structify/plugin/state"
)

// initTemplater is the templater for the init statement.
// It implements the state.Templater interface.
// Add connection to the template
type initTemplater struct {
	state *statepkg.State

	// is include connection
	IncludeConnection bool
}

// NewInitTemplater returns a new initTemplater.
func NewInitTemplater(state *statepkg.State) statepkg.Templater {
	return &initTemplater{
		state: state,

		IncludeConnection: state.IncludeConnection,
	}
}

// BuildTemplate builds the template.
func (i *initTemplater) BuildTemplate() string {
	tmpl, err := helperpkg.ExecuteTemplate(
		tmplpkg.InitStatementTemplate,
		i.Funcs(),
		i,
		helperpkg.IncludeTemplate{
			Name: "connection",
			Body: tmplpkg.ConnectionTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "storages",
			Body: tmplpkg.StorageTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "types",
			Body: tmplpkg.TypesTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "errors",
			Body: tmplpkg.ErrorsTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "repeatedTypes",
			Body: tmplpkg.SingleRepeatedTypesTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "transaction",
			Body: tmplpkg.TransactionManagerTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "options",
			Body: tmplpkg.OptionsTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "conditions",
			Body: tmplpkg.TableConditionsTemplate,
		},
	)
	if err != nil {
		log.Fatalf("failed to execute template: %v", err)
		return ""
	}

	return tmpl
}

// Imports returns the imports.
func (i *initTemplater) Imports() importpkg.ImportSet {
	is := importpkg.ImportSet{}
	is.Enable(
		importpkg.ImportDb,
		importpkg.ImportLibPQ,
		importpkg.ImportStrconv,
		importpkg.ImportFMT,
		importpkg.ImportErrors,
		importpkg.ImportJson,
		importpkg.ImportSQLDriver,
		importpkg.ImportLibPQWOAlias,
		importpkg.ImportStrings,
	)

	return is
}

// Funcs returns the template functions.
func (i *initTemplater) Funcs() map[string]interface{} {
	return template.FuncMap{

		// singleTypesList returns the single types list.
		"singleTypesList": func() statepkg.SingleTypes {
			return i.state.SingleTypes
		},

		// storageNames returns the storage names.
		"storages": func() map[string]string {
			var storages = make(map[string]string)
			for _, m := range i.state.Messages {
				storages[helperpkg.LowerCamelCase(m.GetName())+StoragePostfix] = helperpkg.UpperCamelCase(m.GetName()) + StoragePostfix
			}
			return storages
		},

		// storageName returns the upper camel case storage name. Storages are plural.
		"storageName": func() string {
			return fmt.Sprintf("%s%s", helperpkg.UpperCamelCase(i.state.FileName), helperpkg.UpperCamelCase(helperpkg.Plural(StoragePostfix)))
		},

		// clientName returns the upper camel case client name.
		"clientName": func() string {
			return fmt.Sprintf("%s%s", helperpkg.UpperCamelCase(i.state.FileName), DBClientPostfix)
		},

		// clientName_lower returns the lower camel case client name.
		"clientName_lower": func() string {
			return fmt.Sprintf("%s%s", helperpkg.LowerCamelCase(i.state.FileName), DBClientPostfix)
		},

		// camelCase returns the upper camel case.
		"camelCase": helperpkg.UpperCamelCase,

		// repeat returns the repeated string.
		"repeat": func(string string) string {
			return fmt.Sprintf("%s%s", string, "Repeated")
		},

		// lowerCamelCase returns the lower camel case.
		"lowerCamelCase": helperpkg.LowerCamelCase,

		// nestedMessages returns the nested messages.
		"nestedMessages": func() statepkg.NestedMessages {
			return i.state.NestedMessages
		},

		// messages returns the messages.
		"messages": func() statepkg.Messages {
			return i.state.Messages
		},

		// singleTypes returns the single types.
		"singleTypes": func() statepkg.SingleTypes {
			return i.state.SingleTypes
		},

		// fieldName returns the upper camel case field name.
		"fieldName": func(f *descriptorpb.FieldDescriptorProto) string {
			return helperpkg.UpperCamelCase(f.GetName())
		},

		// isRelation returns the field type.
		"isRelation": func(f *descriptorpb.FieldDescriptorProto) bool {
			return i.state.IsRelation(f)
		},

		// isJSON returns the field type.
		"isJSON": func(f *descriptorpb.FieldDescriptorProto) bool {
			return i.state.NestedMessages.IsJSON(f)
		},

		// fieldType returns the field type.
		"fieldType": func(f *descriptorpb.FieldDescriptorProto) string {
			// if the field is a nested message, return the structure name.
			if i.state.NestedMessages.IsJSON(f) {
				md := i.state.NestedMessages.GetByFieldDescriptor(f)
				if md != nil {
					return helperpkg.TypePrefix(f, md.StructureName)
				}
			}

			return helperpkg.ConvertType(f)
		},

		// sourceName returns the source name.
		"sourceName": func(f *descriptorpb.FieldDescriptorProto) string {
			return f.GetName()
		},
	}
}

// DBClientPostfix is the postfix for the client name.
const DBClientPostfix = "DatabaseClient"
const StoragePostfix = "Storage"
