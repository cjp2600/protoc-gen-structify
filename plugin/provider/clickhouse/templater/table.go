package templater

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"

	importpkg "github.com/cjp2600/protoc-gen-structify/plugin/import"
	helperpkg "github.com/cjp2600/protoc-gen-structify/plugin/pkg/helper"
	tmplpkg "github.com/cjp2600/protoc-gen-structify/plugin/provider/clickhouse/tmpl"
	statepkg "github.com/cjp2600/protoc-gen-structify/plugin/state"
)

// tableTemplater is the templater for the init statement.
// It implements the state.Templater interface.
type tableTemplater struct {
	state   *statepkg.State
	message *descriptorpb.DescriptorProto

	// initMethods bool
	CRUDSchemas bool
}

// NewTableTemplater returns a new initTemplater.
func NewTableTemplater(message *descriptorpb.DescriptorProto, state *statepkg.State) statepkg.Templater {
	return &tableTemplater{
		state:   state,
		message: message,

		CRUDSchemas: state.CRUDSchemas,
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
		helperpkg.IncludeTemplate{
			Name: "create_method",
			Body: tmplpkg.TableCreateMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "async_create_method",
			Body: tmplpkg.TableCreateAsyncMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "batch_create_method",
			Body: tmplpkg.TableBatchCreateMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "original_batch_create_method",
			Body: tmplpkg.TableOriginalBatchCreateMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "update_method",
			Body: tmplpkg.TableUpdateMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "delete_method",
			Body: tmplpkg.TableDeleteMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "raw_method",
			Body: tmplpkg.TableRawQueryMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "get_by_id_method",
			Body: tmplpkg.TableGetByIDMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "find_many_method",
			Body: tmplpkg.TableFindManyMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "find_one_method",
			Body: tmplpkg.TableFindOneMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "count_method",
			Body: tmplpkg.TableCountMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "find_with_pagination",
			Body: tmplpkg.TableFindWithPaginationMethodTemplate,
		},
		helperpkg.IncludeTemplate{
			Name: "table_conditions",
			Body: tmplpkg.TableConditionFilters,
		},
		helperpkg.IncludeTemplate{
			Name: "lock_method",
			Body: tmplpkg.TableLockMethodTemplate,
		},
	)
	if err != nil {
		log.Fatalf("failed to execute template: %v", err)
		return ""
	}

	return tmpl
}

func (t *tableTemplater) TemplateName() string {
	if opts := helperpkg.GetMessageOptions(t.message); opts != nil {
		if opts.Table != "" {
			return opts.Table
		}
	}
	return helperpkg.Plural(t.message.GetName())
}

// Imports returns the imports.
func (t *tableTemplater) Imports() *importpkg.ImportSet {
	is := importpkg.NewImportSet()
	is.Enable(
		importpkg.ImportContext,
		importpkg.ImportFMT,
		importpkg.ImportSquirrel,
		importpkg.ImportClickhouseDriver,
		importpkg.ImportStrings,
	)

	tmp := t.BuildTemplate()
	if strings.Contains(tmp, "time.Time") {
		is.Add(importpkg.ImportTime)
	}
	if strings.Contains(tmp, "null.") {
		is.Add(importpkg.ImportNull)
	}

	return is
}

// Funcs returns the template functions.
func (t *tableTemplater) Funcs() map[string]interface{} {
	return template.FuncMap{

		// isRepeated returns true if the field is repeated.
		"isRepeated": func(f *descriptorpb.FieldDescriptorProto) bool {
			return helperpkg.IsRepeated(f)
		},

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

		"fieldTypeWP": func(f *descriptorpb.FieldDescriptorProto) string {
			// if the field is a single type, return the single type.
			if t.state.SingleTypes.ExistByName(f.GetName()) {
				mds := t.state.SingleTypes.GetByName(f.GetName())
				if mds != nil {
					fieldType := mds.FieldType
					if helperpkg.IsOptional(f) {
						if strings.Contains(mds.FieldType, "*") {
							fieldType = strings.Replace(mds.FieldType, "*", "", 1)
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

			ct := helperpkg.ConvertType(f)

			if helperpkg.IsOptional(f) {
				return strings.Replace(ct, "*", "", 1)
			}

			return ct
		},

		"fieldTypeToNullType": func(f *descriptorpb.FieldDescriptorProto) string {
			// Check if the field is a single type and convert to null types if applicable.
			if t.state.SingleTypes.ExistByName(f.GetName()) {
				mds := t.state.SingleTypes.GetByName(f.GetName())
				if mds != nil {
					fieldType := mds.FieldType

					// Convert basic Go types to their null equivalents.
					switch fieldType {
					case "string":
						return "null.String"
					case "int", "int32", "int64":
						return "null.Int"
					case "float32", "float64":
						return "null.Float"
					case "bool":
						return "null.Bool"
					case "[]byte":
						return "null.Bytes"
					}
				}
			}

			// Check if the field is a nested message and convert accordingly.
			if t.state.NestedMessages.IsJSON(f) {
				md := t.state.NestedMessages.GetByFieldDescriptor(f)
				if md != nil {
					return "NullableJSON[" + helperpkg.TypePrefix(f, md.StructureName) + "]"
				}
			}

			// Default conversion using helper package for other types.
			return helperpkg.ConvertToNullType(f)
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

		// messages returns the messages.
		"messages": func() statepkg.Messages {
			return statepkg.Messages{t.message}
		},

		// messages returns the messages.
		"messages_for_filter": func() statepkg.Messages {
			newMess := helperpkg.CopyMessage(t.message)
			var fields []*descriptorpb.FieldDescriptorProto
			for _, f := range newMess.GetField() {
				opts := helperpkg.GetFieldOptions(f)
				if opts != nil {
					if opts.GetPrimaryKey() ||
						opts.GetInFilter() ||
						t.state.Relations.FindBy(f) ||
						t.state.Relations.FindByMessage(t.message, f) {
						fields = append(fields, f)
					}
				}
			}
			newMess.Field = fields
			return statepkg.Messages{newMess}
		},

		// isJSON returns the field type.
		"isJSON": func(f *descriptorpb.FieldDescriptorProto) bool {
			return t.state.NestedMessages.IsJSON(f)
		},

		// isLastField returns true if the field is the last field.
		"isLastField": func(f *descriptorpb.FieldDescriptorProto) bool {
			var fields []*descriptorpb.FieldDescriptorProto
			for _, f := range t.message.GetField() {
				if !t.state.IsRelation(f) {
					fields = append(fields, f)
				}
			}

			return f == fields[len(fields)-1]
		},

		"isValidLike": func(f *descriptorpb.FieldDescriptorProto) bool {
			switch *f.Type {
			case descriptorpb.FieldDescriptorProto_TYPE_STRING:
				return true
			}
			return false
		},

		"isValidNull": func(f *descriptorpb.FieldDescriptorProto) bool {
			if f == nil {
				return false
			}
			if opts := helperpkg.GetFieldOptions(f); opts != nil {
				return opts.GetNullable()
			}
			return false
		},

		"isValidGT": func(f *descriptorpb.FieldDescriptorProto) bool {
			if f == nil {
				return false
			}
			switch *f.Type {
			case descriptorpb.FieldDescriptorProto_TYPE_INT32:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_INT64:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_STRING:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				parts := strings.Split(f.GetTypeName(), ".")
				typName := parts[len(parts)-1]
				if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
					return true
				}
			}
			return false
		},

		// isValidEq returns the field type.
		"isValidEq": func(f *descriptorpb.FieldDescriptorProto) bool {
			switch *f.Type {
			case descriptorpb.FieldDescriptorProto_TYPE_STRING:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_INT32:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_INT64:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
				return true
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				parts := strings.Split(f.GetTypeName(), ".")
				typName := parts[len(parts)-1]
				if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
					return true
				}
			}

			return false
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

		"hasPrimaryKey": func() bool {
			for _, f := range t.message.GetField() {
				if opts := helperpkg.GetFieldOptions(f); opts != nil {
					if opts.GetPrimaryKey() {
						return true
					}
				}
			}
			return false
		},

		"getPrimaryKey": func() *descriptorpb.FieldDescriptorProto {
			for _, f := range t.message.GetField() {
				if opts := helperpkg.GetFieldOptions(f); opts != nil {
					if opts.GetPrimaryKey() {
						return f
					}
				}
			}
			return nil
		},

		// isPointer returns true if the field is pointer.
		"findPointer": func(f *descriptorpb.FieldDescriptorProto) bool {
			return helperpkg.IsOptional(f)
		},

		// isPrimaryKey returns true if the field is primary key.
		"isPrimary": func(f *descriptorpb.FieldDescriptorProto) bool {
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

		// isDefaultUUID returns true if the field is default uuid.
		"isDefaultUUID": func(f *descriptorpb.FieldDescriptorProto) bool {
			if opts := helperpkg.GetFieldOptions(f); opts != nil {
				if strings.Contains(opts.GetDefault(), "uuid_generate") {
					return true
				}
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

		"isHasRepeated": func() bool {
			for _, f := range t.message.GetField() {
				if helperpkg.IsRepeated(f) {
					return true
				}
			}
			return false
		},

		// storageName returns the upper camel case storage name.
		"storageName": func() string {
			return fmt.Sprintf("%sStorage", helperpkg.UpperCamelCase(t.message.GetName()))
		},

		// messageName returns the upper camel case message name.
		"messageName": func() string {
			return helperpkg.UpperCamelCase(t.message.GetName())
		},

		"message": func() *descriptorpb.DescriptorProto {
			return t.message
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

		// relationAllowSubDeleting returns true if the relation allows sub deleting.
		"relationAllowSubDeleting": func(f *descriptorpb.FieldDescriptorProto) bool {
			return true
		},

		// isRelation returns the field type.
		"isRelation": func(f *descriptorpb.FieldDescriptorProto) bool {
			return t.state.IsRelation(f)
		},

		"isCurrentOptional": func(f *descriptorpb.FieldDescriptorProto) bool {
			return helperpkg.IsOptional(f)
		},

		// isOptional returns true if the field is marked as optional.
		"isOptional": func(f *descriptorpb.FieldDescriptorProto) bool {
			// Construct the relation name based on the message and the field type.
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))

			// Retrieve the relation from the state using the constructed name.
			relation, ok := t.state.Relations.Get(relName)
			if !ok {
				return false
			}

			pd := relation.ParentDescriptor
			for _, fld := range pd.GetField() {
				if opts := helperpkg.GetFieldOptions(f); opts != nil {
					if opts.GetRelation().Field != "" {
						if opts.GetRelation().Field == fld.GetName() {
							return helperpkg.IsOptional(fld)
						}
					}
				}
			}

			// Check if the relation descriptor is marked as optional.
			return false
		},

		// hasRelationOptions returns true if the field has relation options.
		"hasRelationOptions": func(f *descriptorpb.FieldDescriptorProto) bool {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
			_, ok := t.state.Relations.Get(relName)

			if ok {
				opts := helperpkg.GetFieldOptions(f)
				if opts != nil {
					return opts.Relation != nil
				}
			}
			return false
		},

		// hasIndex returns true if the field has index.
		"hasIndex": helperpkg.HasIndex,

		// hasUnique returns true if the field has unique.
		"hasUnique": helperpkg.HasUnique,

		// hasRelation returns true if the message has relation.
		"hasRelation": func() bool {
			for _, f := range t.message.GetField() {
				relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
				_, ok := t.state.Relations.Get(relName)
				if ok {
					return true
				}
			}
			return false
		},

		// relation returns the relation.
		"relation": func(f *descriptorpb.FieldDescriptorProto) *statepkg.Relation {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
			relation, ok := t.state.Relations.Get(relName)
			if !ok {
				return nil
			}

			return relation
		},

		// relationName returns the relation name.
		"relationStorageName": func(f *descriptorpb.FieldDescriptorProto) string {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				return relation.Store
			}
			return ""
		},

		"relationStructureName": func(f *descriptorpb.FieldDescriptorProto) string {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				return relation.StructName
			}
			return ""
		},

		"relationTableName": func(f *descriptorpb.FieldDescriptorProto) string {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				relMess := relation.RelationDescriptor
				if opts := helperpkg.GetMessageOptions(relMess); opts != nil {
					if opts.Table != "" {
						return opts.Table
					}
				}
				return helperpkg.Plural(relMess.GetName())
			}
			return ""
		},

		// relationName returns the relation name.
		"hasIDFromRelation": func(f *descriptorpb.FieldDescriptorProto) bool {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				rd := relation.RelationDescriptor
				for _, f := range rd.GetField() {
					if f.GetName() == "id" {
						return true
					}
				}
			}
			return false
		},

		"getStructureUniqueIndexes": func() map[int][]*descriptorpb.FieldDescriptorProto {
			var indexes = make(map[int][]*descriptorpb.FieldDescriptorProto)
			if opts := helperpkg.GetMessageOptions(t.message); opts != nil {
				if opts.GetUniqueIndex() != nil {
					for indexID, uniqueIndex := range opts.GetUniqueIndex() {
						var fields []*descriptorpb.FieldDescriptorProto
						for _, fieldName := range uniqueIndex.Fields {
							for _, field := range t.message.GetField() {
								if field.GetName() == fieldName {
									fields = append(fields, field)
								}
							}
						}
						if len(fields) > 0 {
							indexes[indexID] = fields
						}
					}
				}
			}
			return indexes
		},

		"sub": func(a, b int) int {
			return a - b
		},

		"sliceToString": func(fields []*descriptorpb.FieldDescriptorProto) string {
			var slice []string
			for _, f := range fields {
				slice = append(slice, helperpkg.SnakeCase(f.GetName()))
			}
			return strings.Join(slice, "_")
		},

		"getFieldID": func(fl *descriptorpb.FieldDescriptorProto) string {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(fl))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				rd := relation.RelationDescriptor
				pd := relation.ParentDescriptor

				if opts := helperpkg.GetFieldOptions(fl); opts != nil {
					if opts.GetRelation().Field != "" {
						return helperpkg.UpperCamelCase(opts.GetRelation().Field)
					}
				}

				if relation.UseTag {
					return helperpkg.UpperCamelCase(relation.Field)
				}

				var currentPrimaryKey string
				for _, f := range pd.GetField() {
					if opts := helperpkg.GetFieldOptions(f); opts != nil {
						if opts.GetPrimaryKey() {
							currentPrimaryKey = helperpkg.UpperCamelCase(f.GetName())
						}
					}
				}

				if helperpkg.DetermineRelationDirection(rd, pd) == "child-to-parent" {
					return currentPrimaryKey
				} else {
					return helperpkg.UpperCamelCase(strings.ToLower(rd.GetName()) + "_id")
				}
			}
			return ""
		},

		"getRefID": func(fl *descriptorpb.FieldDescriptorProto) string {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(fl))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				rd := relation.RelationDescriptor
				pd := relation.ParentDescriptor

				if relation.UseTag {
					return helperpkg.UpperCamelCase(relation.Reference)
				}

				// todo: check

				if helperpkg.DetermineRelationDirection(rd, pd) == "child-to-parent" {
					for _, f := range rd.GetField() {
						if f.GetName() == strings.ToLower(pd.GetName())+"_id" {
							return helperpkg.UpperCamelCase(f.GetName())
						}
					}
				} else {
					for _, f := range rd.GetField() {
						if f.GetName() == "id" {
							return helperpkg.UpperCamelCase(f.GetName())
						}
					}
				}
			}
			return ""
		},

		"getRefSource": func(fl *descriptorpb.FieldDescriptorProto) string {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(fl))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				rd := relation.RelationDescriptor
				pd := relation.ParentDescriptor

				if relation.UseTag {
					return helperpkg.SnakeCase(relation.Reference)
				}

				// todo: check

				if helperpkg.DetermineRelationDirection(rd, pd) == "child-to-parent" {
					for _, f := range rd.GetField() {
						if f.GetName() == strings.ToLower(pd.GetName())+"_id" {
							return helperpkg.SnakeCase(f.GetName())
						}
					}
				} else {
					for _, f := range rd.GetField() {
						if f.GetName() == "id" {
							return helperpkg.SnakeCase(f.GetName())
						}
					}
				}
			}
			return ""
		},

		"isForeign": func(fl *descriptorpb.FieldDescriptorProto) bool {
			if opts := helperpkg.GetFieldOptions(fl); opts != nil {
				if opts.GetRelation() != nil {
					return opts.GetRelation().Foreign != nil
				}
			}
			return false
		},

		"isCascade": func(fl *descriptorpb.FieldDescriptorProto) bool {
			if opts := helperpkg.GetFieldOptions(fl); opts != nil {
				if opts.GetRelation() != nil && opts.GetRelation().Foreign != nil {
					return opts.GetRelation().Foreign.Cascade
				}
			}
			return false
		},

		"getFieldSource": func(fl *descriptorpb.FieldDescriptorProto) string {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(fl))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				rd := relation.RelationDescriptor
				pd := relation.ParentDescriptor

				if relation.UseTag {
					return helperpkg.SnakeCase(relation.Field)
				}

				var currentPrimaryKey string
				for _, f := range pd.GetField() {
					if opts := helperpkg.GetFieldOptions(f); opts != nil {
						if opts.GetPrimaryKey() {
							currentPrimaryKey = helperpkg.SnakeCase(f.GetName())
						}
					}
				}

				if helperpkg.DetermineRelationDirection(rd, pd) == "child-to-parent" {
					return currentPrimaryKey
				} else {
					return helperpkg.SnakeCase(strings.ToLower(rd.GetName()) + "_id")
				}
			}
			return ""
		},

		// relationName returns the relation name.
		"hasID": func() bool {
			for _, f := range t.message.GetField() {
				if f.GetName() == "id" {
					return true
				}
			}
			return false
		},

		"IDType": func() string {
			for _, f := range t.message.GetField() {
				if f.GetName() == "id" {
					return helperpkg.ConvertType(f)
				}
			}
			return "int64"
		},

		// relationAllowSubCreating returns true if the relation allows sub creating.
		"relationAllowSubCreating": func(f *descriptorpb.FieldDescriptorProto) bool {
			relName := t.message.GetName() + "::" + helperpkg.ClearPointer(helperpkg.ConvertType(f))
			relation, ok := t.state.Relations.Get(relName)

			if ok {
				return relation.AllowSubCreating
			}
			return false
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

		"pluralFieldName": func(f *descriptorpb.FieldDescriptorProto) string {
			if helperpkg.IsRepeated(f) {
				return helperpkg.UpperCamelCase(helperpkg.Plural(f.GetName()))
			}
			return helperpkg.UpperCamelCase(f.GetName())
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

		// plural returns the plural.
		"plural": helperpkg.Plural,

		// lowerCamelCase returns the lower camel case.
		"lowerCamelCase": helperpkg.LowerCamelCase,
	}
}
