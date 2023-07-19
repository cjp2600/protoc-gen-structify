package plugin

import (
	"fmt"
	"go/format"
	"html/template"
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

// getMessages returns all the messages in the request. It filters out google.protobuf and structify messages.
func getMessages(req *plugingo.CodeGeneratorRequest) []*descriptorpb.DescriptorProto {
	var messages []*descriptorpb.DescriptorProto

	for _, f := range req.GetProtoFile() {
		for _, m := range f.GetMessageType() {
			if !isUserMessage(f, m) {
				continue
			}
			messages = append(messages, m)
		}
	}

	return messages
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

// goTypeToPostgresType converts a Go type to a Postgres type.
func goTypeToPostgresType(goType string) string {
	switch goType {
	case "string":
		return "TEXT"
	case "bool":
		return "BOOLEAN"
	case "int":
		return "INTEGER"
	case "int32":
		return "INTEGER"
	case "int64":
		return "BIGINT"
	case "float32":
		return "REAL"
	case "float64":
		return "DOUBLE PRECISION"
	case "[]byte":
		return "BYTEA"
	default:
		// This will handle all other types as TEXT
		// You may want to expand this switch to handle other types correctly.
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
		typ = "struct" // This is used for nested messages and it should be handled separately.
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

	return typ
}

// protoToPostgresType converts a protobuf type to a Postgres type.
func protoToPostgresType(fieldType descriptorpb.FieldDescriptorProto_Type) string {
	switch fieldType {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "REAL"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "BIGINT"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "INTEGER"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "BOOLEAN"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING, descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "TEXT"
	default:
		return "TEXT"
	}
}

// isRepeated returns true if the field is repeated.
func isRepeated(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

// Is this field optional?
func isOptional(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_OPTIONAL
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

func upperClientName(name string) string {
	return fmt.Sprintf("%sDB", sToCml(name))
}

func lowerClientName(name string) string {
	return fmt.Sprintf("%sDB", sToLowerCamel(name))
}
