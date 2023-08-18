package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
	"reflect"
	"strings"
	"text/template"
	"unicode"

	"github.com/gertd/go-pluralize"
	"github.com/golang/protobuf/proto"

	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/types/descriptorpb"

	structify "github.com/cjp2600/structify/plugin/options"
)

type DescriptorMList map[string]*descriptorpb.DescriptorProto

func GetFieldOptions(f *descriptorpb.FieldDescriptorProto) *structify.StructifyFieldOptions {
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

// GetMessageOptions returns the custom options for a message.
func GetMessageOptions(d *descriptorpb.DescriptorProto) *structify.StructifyMessageOptions {
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

// GetDBOptions returns the custom options for a file.
func GetDBOptions(f *descriptorpb.FileDescriptorProto) *structify.StructifyDBOptions {
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

// IsUserMessage returns true if the message is not a google.protobuf or structify message.
func IsUserMessage(f *descriptorpb.FileDescriptorProto, m *descriptorpb.DescriptorProto) bool {
	if f.GetPackage() == "google.protobuf" || f.GetPackage() == "structify" {
		return false
	}

	return true
}

// UpperCamelCase converts a string to a CamelCase string.
func UpperCamelCase(name string) string {
	return strcase.UpperCamelCase(name)
}

// LowerCamelCase converts a string to a lowerCamelCase string.
func LowerCamelCase(name string) string {
	return strcase.LowerCamelCase(name)
}

func ToLower(name string) string {
	return strings.ToLower(name)
}

func Plural(name string) string {
	client := pluralize.NewClient()
	plural := client.Plural(name)
	return strings.ToLower(plural)
}

func PostgresType(goType string, options *structify.StructifyFieldOptions, isJson bool) string {
	t := GoTypeToPostgresType(goType)

	// Check if it is a JSON/UUID field
	if options != nil {
		if options.Uuid {
			return "UUID"
		}
		if options.Json {
			return "JSONB"
		}
	}

	if isJson {
		return "JSONB"
	}

	return t
}

func GoTypeToPostgresType(goType string) string {
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
	// TODO: Add cases for other singleTypes as needed
	default:
		return "TEXT"
	}
}

type IncludeTemplate struct {
	Name string
	Body string
}

func ExecuteTemplate(tmpl string, funcs template.FuncMap, data any, templates ...IncludeTemplate) (string, error) {
	var output bytes.Buffer

	t, err := template.New("init").Funcs(funcs).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse init template: %w", err)
	}

	if len(templates) > 0 {
		for _, v := range templates {
			_, err = t.New(v.Name).Parse(v.Body)
			if err != nil {
				panic(err)
			}
		}
	}

	if err = t.Execute(&output, data); err != nil {
		return "", fmt.Errorf("failed to execute init template: %w", err)
	}

	return output.String(), nil
}

func ClearPointer(s string) string {
	s = strings.ReplaceAll(s, "[]", "")
	s = strings.ReplaceAll(s, "*", "")
	return s
}

func IsDefType(field *descriptorpb.FieldDescriptorProto) bool {
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return true
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		rtyp := field.GetTypeName()
		parts := strings.Split(rtyp, ".")
		typName := parts[len(parts)-1]
		if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
			return true
		} else {
			return false
		}
	}

	return false
}

// ConvertType converts a protobuf type to a Go type.
func ConvertType(field *descriptorpb.FieldDescriptorProto) string {
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
			typ = "*" + UpperCamelCase(typName)
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

	if IsRepeated(field) {
		typ = "[]" + typ
	}

	if IsOptional(field) {
		if !strings.Contains(typ, "*") {
			typ = "*" + typ
		}
	}

	return typ
}

func TypePrefix(field *descriptorpb.FieldDescriptorProto, typeName string) string {
	if IsRepeated(field) {
		typeName = "[]" + typeName
	}
	if IsOptional(field) {
		if !strings.Contains(typeName, "*") {
			typeName = "*" + typeName
		}
	}
	return typeName
}

// IsRepeated returns true if the field is repeated.
func IsRepeated(field *descriptorpb.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
}

// IsOptional returns true if the field is optional and not a string, bytes, int32, int64, float32, float64, bool, uint32, uint64 type or a Google Protobuf wrapper message.
func IsOptional(field *descriptorpb.FieldDescriptorProto) bool {
	if field.GetProto3Optional() {
		return true
	}

	if field.Label != nil && *field.Label == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL {
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
			if len(parts) > 2 && parts[len(parts)-2] == "protobuf" {
				if len(parts) > 3 && parts[len(parts)-3] == "google" {
					return false
				}
			}
		}
		return true
	}
	return false
}

// GoFmt formats the generated Go code.
func GoFmt(resp *plugingo.CodeGeneratorResponse) error {
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

// FirstLetterLower converts the first letter of a string to lowercase.
func FirstLetterLower(s string) (string, error) {
	if len(s) == 0 {
		return "", fmt.Errorf("string is empty")
	}

	firstRune := []rune(s)[0]
	return string(unicode.ToLower(firstRune)), nil
}

// SliceToString converts a slice of strings to a string.
func SliceToString(slice []string) string {
	quoted := make([]string, len(slice))
	for i, elem := range slice {
		quoted[i] = fmt.Sprintf("\"%s\"", elem)
	}
	return fmt.Sprintf("[]string{%s}", strings.Join(quoted, ", "))
}

// UpperClientName returns the upperCamelCase name of the client.
func UpperClientName(name string) string {
	return fmt.Sprintf("%sDBClient", UpperCamelCase(name))
}

// LowerClientName returns the lowerCamelCase name of the client.
func LowerClientName(name string) string {
	return fmt.Sprintf("%sDBClient", LowerCamelCase(name))
}

// DetectTableName returns the postgres type for the given type.
func DetectTableName(t string) string {
	name := strings.ReplaceAll(t, "*", "")
	name = strings.ReplaceAll(name, "[]", "")
	return Plural(name)
}

// DetectStoreName returns the postgres type for the given type.
func DetectStoreName(t string) string {
	name := strings.ReplaceAll(t, "*", "")
	name = strings.ReplaceAll(name, "[]", "")
	return UpperCamelCase(name) + "Storage"
}

// DetectStructName returns the struct name for the given type.
func DetectStructName(t string) string {
	name := strings.ReplaceAll(t, "*", "")
	name = strings.ReplaceAll(name, "[]", "")
	return UpperCamelCase(name)
}

// CheckProtoSyntax checks if the syntax of the file is proto3.
func CheckProtoSyntax(file *descriptorpb.FileDescriptorProto) error {
	if file.GetSyntax() != "proto3" {
		return fmt.Errorf("unsupported protobuf syntax: %s, only 'proto3' is supported", file.GetSyntax())
	}

	return nil
}

func DC(f string, to string, s ...interface{}) {
	if strings.Contains(f, to) {
		DumpPrint(s...)
	}
}

func dump(s interface{}) string {
	jsonData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	return string(jsonData)
}

func DumpPrint(values ...interface{}) {
	var resp strings.Builder
	resp.WriteString("\n")
	resp.WriteString("\n")
	for _, v := range values {
		resp.WriteString("\n")
		t := reflect.TypeOf(v)
		resp.WriteString(fmt.Sprintf("=== Name: %s, Type: %s\n", t.Name(), t.String()))
		resp.WriteString(fmt.Sprintf("%+v\n", dump(v)))
		resp.WriteString("===\n")
		resp.WriteString("\n")
	}
	panic(resp.String())
}

// GetUserProtoFiles returns the user proto files.
func GetUserProtoFiles(req *plugingo.CodeGeneratorRequest) []*descriptorpb.FileDescriptorProto {
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

func IsContainsStar(s string) bool {
	return strings.Contains(s, "*")
}

// GetUserProtoFile returns the first user proto file.
func GetUserProtoFile(req *plugingo.CodeGeneratorRequest) *descriptorpb.FileDescriptorProto {
	return GetUserProtoFiles(req)[0]
}

// DetectMany returns true if the field is a many relation.
func DetectMany(t string) bool {
	return strings.Contains(t, "[]")
}

// DetectReference returns the reference field name.
func DetectReference(structName string) string {
	return ToLower(structName) + "_id"
}

// DetectField returns the field name.
func DetectField(structName string) string {
	return "id"
}

func CamelCaseSlice(elem []string) string {
	return UpperCamelCase(strings.Join(elem, ""))
}

func BuildJSONTypeName(parentName string, typeName string) string {
	return "JSON" + UpperCamelCase(parentName) + UpperCamelCase(typeName)
}

func DetermineRelationDirection(rd *descriptorpb.DescriptorProto, pd *descriptorpb.DescriptorProto) string {
	for _, f := range rd.GetField() {
		if f.GetName() == strings.ToLower(pd.GetName())+"_id" {
			return "child-to-parent"
		}
	}

	return "parent-to-child"
}
