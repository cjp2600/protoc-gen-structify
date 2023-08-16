package state

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/types/descriptorpb"

	importpkg "github.com/cjp2600/structify/plugin/import"
	helperpkg "github.com/cjp2600/structify/plugin/pkg/helper"
	"github.com/cjp2600/structify/plugin/pkg/version"
)

// State is the state of the plugin.
type State struct {
	Provider          string // Provider is the Provider of the plugin.
	PackageName       string // PackageName is the package name of the plugin.
	FileName          string // FileName is the file name of the plugin.
	Version           string // Version is the Version of the plugin.
	ProtocVersion     string // ProtocVersion is the Version of protoc.
	FileToGenerate    string // FileToGenerate is the file to generate.
	IncludeConnection bool   // IncludeConnection is the flag to include connection in the generated code.

	Imports        importpkg.ImportSet // Imports is the set of Imports.
	Relations      Relations           // Relations is the set of Relations Messages.
	Messages       Messages            // Messages is the set of root Messages.
	NestedMessages NestedMessages      // NestedMessages is the set of nested Messages.

	// SingleTypes is the set of single types. example: type UserNames []string
	// used for generating json statements.
	SingleTypes SingleTypes
}

func NewState(
	request *plugingo.CodeGeneratorRequest,
) *State {
	protoFile := helperpkg.GetUserProtoFile(request)
	nestedMessages := getNestedMessages(request)
	state := &State{
		Provider:    getProvider(request),
		PackageName: protoFile.GetPackage(),
		FileName:    parseFileName(request),

		Imports:        defaultImports(request),
		Messages:       getMessages(request),
		NestedMessages: nestedMessages,
		Relations:      getRelations(request, nestedMessages),
		ProtocVersion:  getProtocVersion(request),
		Version:        version.GetPluginVersion(),
		FileToGenerate: request.GetFileToGenerate()[0],
		SingleTypes:    getSingleTypes(request, nestedMessages),
	}

	return state
}

func getProvider(request *plugingo.CodeGeneratorRequest) string {
	protoFile := helperpkg.GetUserProtoFile(request)
	opts := helperpkg.GetDBOptions(protoFile)
	if opts != nil {
		return opts.GetProvider()
	}
	return ""
}

// defaultImports returns the default Imports.
func defaultImports(request *plugingo.CodeGeneratorRequest) importpkg.ImportSet {
	protoFile := helperpkg.GetUserProtoFile(request)
	var imports = make(importpkg.ImportSet)

	for _, m := range protoFile.GetMessageType() {
		for _, field := range m.GetField() {
			var typ = field.GetTypeName()
			switch *field.Type {
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				parts := strings.Split(typ, ".")
				typName := parts[len(parts)-1]
				if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
					imports.Enable(importpkg.ImportTime)
				}
			}
		}
	}

	return imports
}

// parseFileName returns the file name of the plugin.
func isAllowSubCreating(request *plugingo.CodeGeneratorRequest, msg *descriptor.DescriptorProto, field *descriptor.FieldDescriptorProto) bool {
	ref := helperpkg.DetectReference(msg.GetName())
	relateDesc := findRelatedDescriptor(request, field)
	if relateDesc != nil {
		for _, f := range relateDesc.GetField() {
			if strings.EqualFold(f.GetName(), ref) {
				return true
			}
		}
	}

	return false
}

func findRelatedDescriptor(request *plugingo.CodeGeneratorRequest, field *descriptor.FieldDescriptorProto) *descriptor.DescriptorProto {
	protoFile := helperpkg.GetUserProtoFile(request)
	convertedType := helperpkg.ConvertType(field)
	for _, msg := range protoFile.GetMessageType() {
		if msg.GetName() == helperpkg.ClearPointer(convertedType) {
			return msg
		}
	}
	return nil
}

// getRelation fills the Relations map in the state struct.
func getRelations(request *plugingo.CodeGeneratorRequest, nestSet NestedMessages) Relations {
	protoFile := helperpkg.GetUserProtoFile(request)
	var respRelations = make(map[RelationType]*Relation)

	for _, msg := range protoFile.GetMessageType() {
		for _, field := range msg.GetField() {
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				convertedType := helperpkg.ConvertType(field)

				relation := &Relation{
					ParentDescriptor:   msg,
					Descriptor:         field,
					RelationDescriptor: findRelatedDescriptor(request, field),
					Field:              helperpkg.DetectField(helperpkg.DetectStructName(convertedType)),
					Reference:          helperpkg.DetectReference(msg.GetName()),
					TableName:          helperpkg.DetectTableName(convertedType),  // Assuming msg.GetName() is the table name
					StructName:         helperpkg.DetectStructName(convertedType), // Assuming field.GetName() is the struct name
					Store:              helperpkg.DetectStoreName(convertedType),  // Fill this with the proper value
					Many:               helperpkg.DetectMany(convertedType),       // As the field is repeated, it means there are many Relations
					AllowSubCreating:   isAllowSubCreating(request, msg, field),   // default allow sub creating
					Limit:              100,                                       // default relation limit
				}

				options := helperpkg.GetFieldOptions(field)
				if options != nil {
					relOptions := options.GetRelation()
					if relOptions != nil {
						relation.Field = relOptions.GetField()
						relation.Reference = relOptions.GetReference()
						relation.Limit = uint64(relOptions.GetLimit())
					}
				}

				if nestSet.CheckIsRelation(field) {
					// Add the relation to the map of Relations
					respRelations[NewRelationType(msg.GetName(), relation.StructName)] = relation
				}
			}
		}
	}

	for _, msg := range protoFile.GetMessageType() {
		for _, field := range msg.GetField() {
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				convertedType := helperpkg.ConvertType(field)
				structName := helperpkg.DetectStructName(convertedType)
				if v, ok := respRelations[NewRelationType(structName, msg.GetName())]; ok {
					// If the relation is already in the map, it means that it is a many-to-many relation
					// and we need to fill the reference and field values of the relation.
					// The reference and field values are the opposite of the values of the relation
					// that is already in the map.
					respRelations[NewRelationType(msg.GetName(), structName)].Field = v.Reference
					respRelations[NewRelationType(msg.GetName(), structName)].Reference = v.Field
				}
			}
		}
	}

	return respRelations
}

// GetFlattenNestedMessages checks that the protobuf syntax is supported.
func getNestedMessages(request *plugingo.CodeGeneratorRequest) map[string]*MessageDescriptor {
	result := make(map[string]*MessageDescriptor)
	file := helperpkg.GetUserProtoFile(request)

	for _, msg := range file.GetMessageType() {
		if len(msg.GetNestedType()) == 0 {
			continue
		}

		// find only nested Messages
		// if the message has nested Messages, flatten it
		// and add to the result
		for _, nested := range msg.GetNestedType() {
			flattenMessage(nested, result, msg.GetName()+".")
		}
	}

	// replace the field SingleTypes
	// if the field type is a nested message, replace it with the flattened message
	for _, msgDesc := range result {
		replaceFieldTypes(msgDesc, result)
	}

	return result
}

// flattenMessage - flatten the message
func flattenMessage(msg *descriptor.DescriptorProto, result map[string]*MessageDescriptor, parent string) {
	sourceName := msg.GetName()
	name := helperpkg.CamelCaseSlice(strings.Split(parent+sourceName, "."))
	result[sourceName] = &MessageDescriptor{
		Descriptor:    msg,
		StructureName: name,
		SourceName:    sourceName,
	}

	for _, nested := range msg.GetNestedType() {
		flattenMessage(nested, result, sourceName+".")
	}
}

// replaceFieldTypes - replace the field SingleTypes
func replaceFieldTypes(msgDesc *MessageDescriptor, msgMap map[string]*MessageDescriptor) {
	for _, field := range msgDesc.Descriptor.GetField() {
		if field.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			convertedType := helperpkg.ConvertType(field)
			if fieldDesc, ok := msgMap[helperpkg.DetectStructName(convertedType)]; ok {
				field.TypeName = &fieldDesc.StructureName
			}
		}
	}
}

// getMessages returns the Messages and nested Messages.
func getMessages(request *plugingo.CodeGeneratorRequest) []*descriptorpb.DescriptorProto {
	var messages []*descriptorpb.DescriptorProto

	f := helperpkg.GetUserProtoFile(request)
	for _, m := range f.GetMessageType() {
		if !helperpkg.IsUserMessage(f, m) {
			continue
		}
		messages = append(messages, m)
	}

	return messages
}

// parseFileName parses the file name from the protobuf request.
func parseFileName(request *plugingo.CodeGeneratorRequest) string {
	fileBase := path.Base(request.GetFileToGenerate()[0])
	fileExt := path.Ext(fileBase)
	return strings.TrimSuffix(fileBase, fileExt)
}

// getProtocVersion returns the protoc Version from the protobuf request.
func getProtocVersion(request *plugingo.CodeGeneratorRequest) string {
	ver := request.GetCompilerVersion()
	return fmt.Sprintf("%d.%d.%d", ver.GetMajor(), ver.GetMinor(), ver.GetPatch())
}

// SingleTypes is a type for how to generate json statements.
type SingleTypes map[string]SingleType

// String returns a string representation of the SingleTypes.
func (j SingleTypes) String() string {
	b, err := json.Marshal(j)
	if err != nil {
		return ""
	}

	return string(b)
}

func (j SingleTypes) ExistByName(name string) bool {
	for k, _ := range j {
		if strings.Contains(k, "::"+name) {
			return true
		}
	}
	return false
}

func (j SingleTypes) GetByName(name string) *SingleType {
	for k, v := range j {
		if strings.Contains(k, "::"+name) {
			return &v
		}
	}
	return nil
}

func (j SingleTypes) Get(name string) (SingleType, bool) {
	val, ok := j[name]
	return val, ok
}

func (j SingleTypes) Delete(name string) {
	delete(j, name)
}

// IsExist checks if the given name exists in the Messages.
func (j SingleTypes) IsExist(f *descriptorpb.FieldDescriptorProto) bool {
	for k, v := range j {
		for _, n := range []string{
			f.GetName(),
			f.GetTypeName(),
			helperpkg.CamelCaseSlice(strings.Split(f.GetTypeName(), ".")),
			helperpkg.DetectStructName(helperpkg.ConvertType(f)),
			helperpkg.ConvertType(f),
		} {
			if strings.EqualFold(v.Descriptor.GetName(), n) {
				return true
			}
			if strings.EqualFold(k, n) {
				return true
			}
		}
	}
	return false
}

func (j SingleTypes) GetByFieldDescriptor(f *descriptorpb.FieldDescriptorProto) *SingleType {
	for k, v := range j {
		for _, n := range []string{
			f.GetName(),
			f.GetTypeName(),
			helperpkg.CamelCaseSlice(strings.Split(f.GetTypeName(), ".")),
			helperpkg.DetectStructName(helperpkg.ConvertType(f)),
			helperpkg.ConvertType(f),
		} {
			if strings.EqualFold(v.Descriptor.GetName(), n) {
				return &v
			}
			if strings.EqualFold(k, n) {
				return &v
			}
		}
	}

	return nil
}

type Relations map[RelationType]*Relation

func (r Relations) String() string {
	var output string
	for k := range r {
		output += k.String() + "\n"
	}
	return output
}

func (r Relations) Delete(name string) {
	delete(r, RelationType(name))
}

func (r Relations) Get(name string) (*Relation, bool) {
	val, ok := r[RelationType(name)]
	return val, ok
}

// IsExist checks if the given name exists in the Messages.
func (r Relations) IsExist(f *descriptorpb.FieldDescriptorProto) bool {
	for k, v := range r {
		for _, n := range []string{
			f.GetName(),
			f.GetTypeName(),
			helperpkg.CamelCaseSlice(strings.Split(f.GetTypeName(), ".")),
			helperpkg.DetectStructName(helperpkg.ConvertType(f)),
			helperpkg.ConvertType(f),
		} {
			if strings.EqualFold(v.Descriptor.GetName(), n) {
				return true
			}
			if strings.EqualFold(k.String(), n) {
				return true
			}
		}
	}
	return false
}

// GetByFieldDescriptor returns the Relation by the given FieldDescriptorProto.
func (r Relations) GetByFieldDescriptor(f *descriptorpb.FieldDescriptorProto) *Relation {
	for k, v := range r {
		for _, n := range []string{
			f.GetName(),
			f.GetTypeName(),
			helperpkg.CamelCaseSlice(strings.Split(f.GetTypeName(), ".")),
			helperpkg.DetectStructName(helperpkg.ConvertType(f)),
			helperpkg.ConvertType(f),
		} {
			if strings.EqualFold(v.Descriptor.GetName(), n) {
				return v
			}
			if strings.EqualFold(k.String(), n) {
				return v
			}
		}
	}

	return nil
}

type Messages []*descriptorpb.DescriptorProto

func (t Messages) String() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func (t Messages) FindByName(name string) *descriptorpb.DescriptorProto {
	for _, m := range t {
		if m.GetName() == name {
			return m
		}
	}

	return nil
}

type NestedMessages map[string]*MessageDescriptor

// String returns a string representation of the NestedMessages.
func (t NestedMessages) String() string {
	b, err := json.Marshal(t)
	if err != nil {
		return ""
	}

	return string(b)
}

func (t NestedMessages) CheckIsRelation(f *descriptorpb.FieldDescriptorProto) bool {
	if _, ok := t[helperpkg.DetectStructName(helperpkg.ConvertType(f))]; ok {
		return false
	}

	// Check if it is a message type
	if *f.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		// If it is, check if it is a system message type
		typ := f.GetTypeName()
		parts := strings.Split(typ, ".")
		typName := parts[len(parts)-1]

		// Exclude system singleTypes such as google.protobuf.Timestamp
		if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
			return false
		}

		return true
	}

	return false
}

// IsJSON returns true if the field is a JSON field.
func (t NestedMessages) IsJSON(f *descriptorpb.FieldDescriptorProto) bool {
	if t.IsExist(f) {
		return true
	}

	convertedType := helperpkg.ConvertType(f)
	if helperpkg.DetectMany(convertedType) && !t.CheckIsRelation(f) {
		return true
	}

	return false
}

// GetDescriptor returns the table with the given name.
func (t NestedMessages) GetDescriptor(name string) (*descriptorpb.DescriptorProto, bool) {
	for _, v := range t {
		if v.Descriptor.GetName() == name {
			return v.Descriptor, true
		}
	}
	return nil, false
}

// IsExist checks if the given name exists in the Messages.
func (t NestedMessages) IsExist(f *descriptorpb.FieldDescriptorProto) bool {
	for k, v := range t {
		for _, n := range []string{
			f.GetName(),
			f.GetTypeName(),
			helperpkg.CamelCaseSlice(strings.Split(f.GetTypeName(), ".")),
			helperpkg.DetectStructName(helperpkg.ConvertType(f)),
			helperpkg.ConvertType(f),
		} {
			if strings.EqualFold(v.Descriptor.GetName(), n) {
				return true
			}
			if strings.EqualFold(k, n) {
				return true
			}
		}
	}
	return false
}

func (t NestedMessages) GetByField(f *descriptorpb.FieldDescriptorProto) *MessageDescriptor {
	for k, v := range t {
		for _, n := range []string{
			f.GetName(),
			f.GetTypeName(),
			helperpkg.CamelCaseSlice(strings.Split(f.GetTypeName(), ".")),
			helperpkg.DetectStructName(helperpkg.ConvertType(f)),
			helperpkg.ConvertType(f),
		} {
			if strings.EqualFold(v.Descriptor.GetName(), n) {
				return v
			}
			if strings.EqualFold(k, n) {
				return v
			}
		}
	}

	return nil
}

// GetByFieldDescriptor gets the table by the given descriptor.
func (t NestedMessages) GetByFieldDescriptor(f *descriptorpb.FieldDescriptorProto) *MessageDescriptor {
	for k, v := range t {
		for _, n := range []string{
			f.GetName(),
			f.GetTypeName(),
			helperpkg.CamelCaseSlice(strings.Split(f.GetTypeName(), ".")),
			helperpkg.DetectStructName(helperpkg.ConvertType(f)),
			helperpkg.ConvertType(f),
		} {
			if strings.EqualFold(v.Descriptor.GetName(), n) {
				return v
			}
			if strings.EqualFold(k, n) {
				return v
			}
		}
	}

	return nil
}

func (t NestedMessages) Get(name string) *MessageDescriptor {
	if v, ok := t[name]; ok {
		return v
	}

	return nil
}

// IsExistInTables checks if the given name exists in the Messages.
func (s *State) IsExistInTables(name string) bool {
	for _, t := range s.Messages {
		if t.GetName() == name {
			return true
		}
	}
	return false
}

// IsExistInNestedTables checks if the given name exists in the nested Messages.
func (s *State) IsExistInNestedTables(name string) bool {
	for _, t := range s.NestedMessages {
		if t.Descriptor.GetName() == name {
			return true
		}
	}
	return false
}

func (s *State) String() string {
	return "State{" +
		"Provider: " + s.Provider +
		", PackageName: " + s.PackageName +
		", FileName: " + s.FileName +
		", Imports: " + s.Imports.String() +
		", Relations: " + s.Relations.String() +
		", Messages: " + s.Messages.String() +
		", NestedMessages: " + s.NestedMessages.String() +
		", SingleTypes: " + s.SingleTypes.String() +
		"}"
}

// ImportsFromTable Imports the given table.
func (s *State) ImportsFromTable(tables []Templater) {
	for _, t := range tables {
		if t == nil {
			continue
		}

		for i, v := range t.Imports() {
			if v {
				s.Imports[i] = v
			}
		}
	}
}

type SingleType struct {
	ParentDescriptor *descriptor.DescriptorProto
	Descriptor       *descriptor.FieldDescriptorProto

	ParentStructureName string
	FieldName           string
	FieldType           string
	SourceType          string
	SourceName          string
	StructureName       string
	Repeated            bool
}

type NestedTableVal struct {
	StructureName string
	HasType       bool
}

type Relation struct {
	RelationDescriptor *descriptor.DescriptorProto
	ParentDescriptor   *descriptor.DescriptorProto
	Descriptor         *descriptor.FieldDescriptorProto
	Field              string
	Reference          string
	TableName          string
	StructName         string
	Store              string
	Limit              uint64
	Many               bool
	AllowSubCreating   bool
}

type RelationType string

// Relation SingleTypes.
func (r RelationType) String() string {
	return string(r)
}

// NewRelationType creates a new relation type.
func NewRelationType(messageName string, structureName string) RelationType {
	return RelationType(messageName + "::" + structureName)
}

// getSingleTypes returns the SingleTypes.
func getSingleTypes(request *plugingo.CodeGeneratorRequest, messages NestedMessages) SingleTypes {
	file := helperpkg.GetUserProtoFile(request)
	singleTypes := make(map[string]SingleType)

	// Get all the SingleTypes.
	for _, m := range file.GetMessageType() {
		for _, field := range m.GetField() {
			if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				repeated := helperpkg.IsRepeated(field)

				if messages.IsJSON(field) {
					convertedType := helperpkg.ConvertType(field)
					fieldName := getRepeatedFieldName(field, repeated)
					linkedStructName := getLinkedStructName(m, field, messages, convertedType)

					singleTypes[m.GetName()+"::"+field.GetName()] = SingleType{
						ParentDescriptor:    m,
						Descriptor:          field,
						ParentStructureName: m.GetName(),
						FieldName:           fieldName,
						FieldType:           linkedStructName + "Repeated",
						SourceType:          convertedType,
						SourceName:          field.GetName(),
						StructureName:       linkedStructName,
						Repeated:            repeated,
					}
				}
			}
		}
	}

	return singleTypes
}

// getRepeatedFieldName returns the field name of a repeated field.
func getRepeatedFieldName(field *descriptor.FieldDescriptorProto, repeated bool) string {
	fieldName := field.GetName()
	if repeated {
		fieldName += "Repeated"
	}
	return fieldName
}

// getLinkedStructName returns the linked struct name of a repeated field.
func getLinkedStructName(m *descriptor.DescriptorProto, field *descriptor.FieldDescriptorProto, messages NestedMessages, convertedType string) string {
	linkedStructName := helperpkg.UpperCamelCase(m.GetName()) + helperpkg.UpperCamelCase(field.GetName())

	if val, ok := messages[convertedType]; ok {
		linkedStructName = val.Descriptor.GetName()
	}
	return linkedStructName
}

// buildJSONTypeTemplate builds the template of a repeated field.
func buildJSONTypeTemplate(linkedStructName string, convertedType string, repeated bool) string {
	template := fmt.Sprintf(`type %s %s`+"\n", linkedStructName, convertedType)

	if repeated {
		convertedType = "[]*" + linkedStructName
		template = fmt.Sprintf(`type %s %s`+"\n", linkedStructName+"Repeated", convertedType)
	}

	return template
}

// MessageDescriptor is a descriptor for a message.
type MessageDescriptor struct {
	Descriptor    *descriptor.DescriptorProto
	StructureName string
	SourceName    string
}

// Templater is an interface for generating templates.
type Templater interface {
	BuildTemplate() string
	Imports() importpkg.ImportSet
}
