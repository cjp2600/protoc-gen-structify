package plugin

import (
	"encoding/json"
	"fmt"
	"go/format"
	"html/template"
	"log"
	"strings"
	"unicode"

	"github.com/gertd/go-pluralize"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/types/descriptorpb"

	structify "github.com/cjp2600/structify/plugin/options"
)

type DescriptorMList map[string]*descriptorpb.DescriptorProto

func getFieldOptions(f *descriptorpb.FieldDescriptorProto) *structify.StructifyFieldOptions {
	opts := f.GetOptions()
	if opts != nil {
		ext, _ := proto.GetExtension(opts, structify.E_Field)
		if ext != nil {
			customOpts, ok := ext.(*structify.StructifyFieldOptions)
			if ok {
				return customOpts
			}
		}
	}
	return nil
}

// getMessageOptions returns the custom options for a message.
func getMessageOptions(d *descriptorpb.DescriptorProto) *structify.StructifyMessageOptions {
	opts := d.GetOptions()
	if opts != nil {
		ext, _ := proto.GetExtension(opts, structify.E_Opts)
		if ext != nil {
			customOpts, ok := ext.(*structify.StructifyMessageOptions)
			if ok {
				return customOpts
			}
		}
	}
	return nil
}

// getDBOptions returns the custom options for a file.
func getDBOptions(f *descriptorpb.FileDescriptorProto) *structify.StructifyDBOptions {
	opts := f.GetOptions()
	if opts != nil {
		ext, err := proto.GetExtension(opts, structify.E_Db)
		if err == nil && ext != nil {
			if customOpts, ok := ext.(*structify.StructifyDBOptions); ok {
				return customOpts
			}
		}
	}
	return nil
}

// getMessages returns the messages and nested messages.
func getMessages(req *plugingo.CodeGeneratorRequest) ([]*descriptorpb.DescriptorProto, []*descriptorpb.DescriptorProto) {
	var messages []*descriptorpb.DescriptorProto
	var nestedMessages []*descriptorpb.DescriptorProto

	var parentPrefix = "JSON"
	f := getUserProtoFile(req)
	for _, m := range f.GetMessageType() {
		if !isUserMessage(f, m) {
			continue
		}
		messages = append(messages, m)

		// Add nested messages to the separate list
		nestedMessages = append(nestedMessages, getNestedMessages(f, m.GetNestedType(), parentPrefix+m.GetName())...)
	}

	return messages, nestedMessages
}

// getNestedMessages returns the nested messages.
func getNestedMessages(file *descriptorpb.FileDescriptorProto, descriptors []*descriptorpb.DescriptorProto, parentPrefix string) []*descriptorpb.DescriptorProto {
	var messages []*descriptorpb.DescriptorProto

	for _, m := range descriptors {
		if !isUserMessage(file, m) {
			continue
		}

		// Add prefix to message name
		m.Name = proto.String(parentPrefix + m.GetName())

		// Add this message to the list
		messages = append(messages, m)

		// Recursively add nested messages of this message
		messages = append(messages, getNestedMessages(file, m.GetNestedType(), *m.Name+"")...)
	}

	return messages
}

func (d *DescriptorMList) exists(name string) bool {
	_, ok := (*d)[name]
	return ok
}

func (d *DescriptorMList) getDescriptor(name string) *descriptorpb.DescriptorProto {
	return (*d)[name]
}

func (d *DescriptorMList) getDescriptorByType(typ string) *descriptorpb.DescriptorProto {
	for _, v := range *d {
		if v.GetName() == typ {
			return v
		}
	}
	return nil
}

func (d *DescriptorMList) getDescriptorByField(field string) *descriptorpb.DescriptorProto {
	for _, v := range *d {
		for _, f := range v.GetField() {
			if f.GetName() == field {
				return v
			}
		}
	}
	return nil
}

func descriptorToMap(descriptors []*descriptorpb.DescriptorProto) DescriptorMList {
	m := make(map[string]*descriptorpb.DescriptorProto)
	for _, d := range descriptors {
		m[*d.Name] = d
		if len(d.GetNestedType()) > 0 {
			m = mergeDescriptors(m, descriptorToMap(d.GetNestedType()))
		}
	}
	return m
}

func mergeDescriptors(m1, m2 map[string]*descriptorpb.DescriptorProto) map[string]*descriptorpb.DescriptorProto {
	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}

// isUserMessage returns true if the message is not a google.protobuf or structify message.
func isUserMessage(f *descriptorpb.FileDescriptorProto, m *descriptorpb.DescriptorProto) bool {
	if f.GetPackage() == "google.protobuf" || f.GetPackage() == "structify" {
		return false
	}

	return true
}

// sToCml converts a string to a CamelCase string.
func sToCml(name string) string {
	return strcase.UpperCamelCase(name)
}

// sToLowerCamel converts a string to a lowerCamelCase string.
func sToLowerCamel(name string) string {
	return strcase.LowerCamelCase(name)
}

func lowerCase(name string) string {
	return strings.ToLower(name)
}

func lowerCasePlural(name string) string {
	client := pluralize.NewClient()
	plural := client.Plural(name)
	return strings.ToLower(plural)
}

func postgresType(goType string, options *structify.StructifyFieldOptions) string {
	t := goTypeToPostgresType(goType)

	// Check if it is a JSON/UUID field
	if options != nil {
		if options.Uuid {
			return "UUID"
		}
		if options.Json {
			return "JSONB"
		}
	}

	return t
}

func goTypeToPostgresType(goType string) string {
	goType = strings.TrimPrefix(goType, "*")
	switch goType {
	case "string":
		return "TEXT"
	case "bool":
		return "BOOLEAN"
	case "int", "int32":
		return "INTEGER"
	case "int64":
		return "BIGINT"
	case "float32":
		return "REAL"
	case "float64":
		return "DOUBLE PRECISION"
	case "time.Time":
		return "TIMESTAMP"
	case "[]byte":
		return "BYTEA"
	// TODO: Add cases for other types as needed
	default:
		return "TEXT"
	}
}

// convertType converts a protobuf type to a Go type.
func convertType(field *descriptor.FieldDescriptorProto) string {
	var typ = field.GetTypeName()

	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		typ = "float64"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		typ = "float32"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		typ = "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		typ = "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		typ = "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		typ = "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		typ = "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		typ = "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		typ = "string"
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		typ = "error" // Group type is deprecated and not recommended.
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		parts := strings.Split(typ, ".")
		typName := parts[len(parts)-1]
		if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
			typ = "time.Time"
		} else {
			typ = "*" + sToCml(typName)
		}
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		typ = "[]byte"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		typ = "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		typ = "int32" // Enums are represented as integers in Go.
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		typ = "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		typ = "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		typ = "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		typ = "int64"
	}

	if isRepeated(field) {
		typ = "[]" + typ
	}

	if isOptional(field) {
		if !strings.Contains(typ, "*") {
			typ = "*" + typ
		}
	}

	return typ
}

// isRepeated returns true if the field is repeated.
func isRepeated(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

// Is this field optional?
// isOptional returns true if the field is optional and not a string, bytes, int32, int64, float32, float64, bool, uint32, uint64 type or a Google Protobuf wrapper message.
func isOptional(field *descriptor.FieldDescriptorProto) bool {
	if field.GetProto3Optional() {
		return true
	}

	if field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_OPTIONAL {
		switch *field.Type {
		case descriptorpb.FieldDescriptorProto_TYPE_STRING,
			descriptorpb.FieldDescriptorProto_TYPE_BYTES,
			descriptorpb.FieldDescriptorProto_TYPE_INT32,
			descriptorpb.FieldDescriptorProto_TYPE_INT64,
			descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
			descriptorpb.FieldDescriptorProto_TYPE_BOOL,
			descriptorpb.FieldDescriptorProto_TYPE_UINT32,
			descriptorpb.FieldDescriptorProto_TYPE_UINT64:
			return false
		case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			// Check if the type is a Google Protobuf wrapper message.
			var typ = field.GetTypeName()
			parts := strings.Split(typ, ".")
			if parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
				return false
			}
		}
		return true
	}
	return false
}

// Is this field required?
func isRequired(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_REQUIRED
}

// goFmt formats the generated Go code.
func goFmt(resp *plugingo.CodeGeneratorResponse) error {
	for i := 0; i < len(resp.File); i++ {
		formatted, err := format.Source([]byte(resp.File[i].GetContent()))
		if err != nil {
			return fmt.Errorf("go format error: %v", err)
		}

		fmts := string(formatted)
		resp.File[i].Content = &fmts
	}
	return nil
}

// firstLetterLower converts the first letter of a string to lowercase.
func firstLetterLower(s string) (string, error) {
	if len(s) == 0 {
		return "", fmt.Errorf("string is empty")
	}

	firstRune := []rune(s)[0]
	return string(unicode.ToLower(firstRune)), nil
}

// sliceToString converts a slice of strings to a string.
func sliceToString(slice []string) template.HTML {
	quoted := make([]string, len(slice))
	for i, elem := range slice {
		quoted[i] = fmt.Sprintf("\"%s\"", elem)
	}
	return template.HTML(fmt.Sprintf("[]string{%s}", strings.Join(quoted, ", ")))
}

// upperClientName returns the upperCamelCase name of the client.
func upperClientName(name string) string {
	return fmt.Sprintf("%sDBClient", sToCml(name))
}

// lowerClientName returns the lowerCamelCase name of the client.
func lowerClientName(name string) string {
	return fmt.Sprintf("%sDBClient", sToLowerCamel(name))
}

// postgresType returns the postgres type for the given type.
func detectTableName(t string) string {
	name := strings.ReplaceAll(t, "*", "")
	name = strings.ReplaceAll(name, "[]", "")
	return lowerCasePlural(name)
}

// detectStoreName returns the postgres type for the given type.
func detectStoreName(t string) string {
	name := strings.ReplaceAll(t, "*", "")
	name = strings.ReplaceAll(name, "[]", "")
	return sToCml(name) + "Store"
}

// detectStructName returns the struct name for the given type.
func detectStructName(t string) string {
	name := strings.ReplaceAll(t, "*", "")
	name = strings.ReplaceAll(name, "[]", "")
	return sToCml(name)
}

// checkIsRelation checks if the field is a relation.
func checkIsRelation(f *descriptorpb.FieldDescriptorProto) bool {
	// Check if it is a message type
	if *f.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		// If it is, check if it is a system message type
		typ := f.GetTypeName()
		parts := strings.Split(typ, ".")
		typName := parts[len(parts)-1]

		// Exclude system types such as google.protobuf.Timestamp
		if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
			return false
		}

		return true
	}

	return false
}

// checkProtoSyntax checks if the syntax of the file is proto3.
func checkProtoSyntax(file *descriptor.FileDescriptorProto) error {
	if file.GetSyntax() != "proto3" {
		return fmt.Errorf("unsupported protobuf syntax: %s, only 'proto3' is supported", file.GetSyntax())
	}

	return nil
}

func dump(s interface{}) string {
	jsonData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	return string(jsonData)
}

func dumpP(s interface{}) {
	panic(fmt.Sprintf("%+v", dump(s)))
}

func dumpF(s interface{}) {
	fmt.Println("")
	fmt.Println(fmt.Sprintf("%+v", dump(s)))
	fmt.Println("")
	fmt.Println("//------------------------")
}

// getUserProtoFiles returns the user proto files.
func getUserProtoFiles(req *plugingo.CodeGeneratorRequest) []*descriptorpb.FileDescriptorProto {
	var userProtoFiles []*descriptorpb.FileDescriptorProto
	filesToGenerate := make(map[string]bool)
	for _, fileName := range req.GetFileToGenerate() {
		filesToGenerate[fileName] = true
	}

	for _, protoFile := range req.GetProtoFile() {
		if _, ok := filesToGenerate[*protoFile.Name]; ok {
			userProtoFiles = append(userProtoFiles, protoFile)
		}
	}

	return userProtoFiles
}

func isContainsStar(s string) bool {
	return strings.Contains(s, "*")
}

// getUserProtoFile returns the first user proto file.
func getUserProtoFile(req *plugingo.CodeGeneratorRequest) *descriptorpb.FileDescriptorProto {
	return getUserProtoFiles(req)[0]
}

// isJSON returns true if the field is a JSON field.
func isJSON(f *descriptorpb.FieldDescriptorProto, state *State) bool {
	convertedType := convertType(f)
	if _, ok := state.NestedTableStructMapping[convertedType]; ok {
		return true
	}

	if detectMany(convertedType) && !checkIsRelation(f) {
		return true
	}

	return false
}

// detectMany returns true if the field is a many relation.
func detectMany(t string) bool {
	return strings.Contains(t, "[]")
}

// detectReference returns the reference field name.
func detectReference(structName string) string {
	return lowerCase(structName) + "_id"
}

// detectField returns the field name.
func detectField(structName string) string {
	return "id"
}

func buildJSONTypeName(parentName string, typeName string) string {
	return "JSON" + sToCml(parentName) + sToCml(typeName)
}
