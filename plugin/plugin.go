package plugin

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ErrUnsupportedProvider is returned when the provider is not supported.
var ErrUnsupportedProvider = errors.New("unsupported provider")

// Templater is an interface for generating templates.
type Templater interface {
	BuildTemplate() string
	Imports() ImportSet
}

// Provider represents the database provider.
type Provider string

// String returns the provider as a string.
func (p Provider) String() string {
	return string(p)
}

// Available providers.
var (
	ProviderMysql    Provider = "mysql"
	ProviderPostgres Provider = "postgres"
	ProviderSqlite   Provider = "sqlite"
)

// PathType is a type for how to generate output filenames.
type PathType int

// Available path types.
const (
	PathTypeImport PathType = iota
	PathTypeSourceRelative
)

const GeneratedFilePostfix = ".db.go"

// Plugin handles generation of code based on protobufs.
type Plugin struct {
	req *plugingo.CodeGeneratorRequest
	res *plugingo.CodeGeneratorResponse

	pathType PathType
	provider Provider
	state    *State

	Param              map[string]string
	PackageName        string
	FileNameWithoutExt string
}

// NewPlugin creates a new Plugin.
func NewPlugin() *Plugin {
	return &Plugin{
		req: &plugingo.CodeGeneratorRequest{},
		res: &plugingo.CodeGeneratorResponse{},
	}
}

// Run handles the input/output of the plugin.
// It reads the request from stdin and writes the response to stdout.
func (p *Plugin) Run() {
	// read from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read from stdin: %v", err)
	}

	// unmarshal protobuf from stdin to request struct and check for errors
	if err := proto.Unmarshal(data, p.req); err != nil {
		log.Fatalf("Failed to unmarshal protobuf: %v", err)
	}

	// check protobuf version
	if err := p.checkProtobufVersion(); err != nil {
		log.Fatalf("Failed to check protobuf version: %v", err)
	}

	if len(getUserProtoFiles(p.req)) == 0 {
		log.Fatalf("No proto file is supported: %d", len(getUserProtoFiles(p.req)))
	}

	// only one proto file is supported
	if len(getUserProtoFiles(p.req)) > 1 {
		log.Fatalf("Only one proto file is supported: %d", len(getUserProtoFiles(p.req)))
	}

	// fill default state
	//
	p.state = p.fillState()
	{
		p.FileNameWithoutExt = p.state.FileName
		p.PackageName = p.state.PackageName
		p.provider = p.state.Provider
	}

	// fill relations map in state struct
	//		- field name
	//		- reference name
	//		- table name
	//		- struct name
	//		- store name
	//		- many
	p.state = p.fillRelation(p.state)

	// fill nested table struct mapping
	// 		- nested table name
	// 		- struct name
	// 		- has type
	//
	p.state = p.fillNestedTableStructMapping(p.state)

	// fill json types
	// 		- structure name
	// 		- field name
	// 		- field type
	// 		- type name
	// 		- template
	p.state = p.fillJSONTypes(p.state)

	// parse command line parameters
	// 	- paths
	// 	- provider
	// 	- timeout
	//
	p.parseCommandLineParameters(p.req.GetParameter())

	// tables is a slice of Templater interface
	// it contains all the tables that will be generated
	tables, err := p.getTemplaterTables()
	if err != nil {
		log.Fatalf("Failed to parse protobuf: %v", err)
	}

	// nestedTables is a slice of Templater interface
	// it contains all the nested tables that will be generated
	nestedTables, err := p.getNestedTemplater()
	if err != nil {
		log.Fatalf("Failed to parse protobuf: %v", err)
	}

	// fill imports
	//
	// 	- import "database/sql"
	// 	- import "github.com/lib/pq"
	// 	- import "github.com/Masterminds/squirrel"
	// 	- import "context"
	// 	- import "errors"
	// 	- import "fmt"
	// 	- import "strings"
	//
	p.state.FillImports(tables)
	p.EnableDefaultImports()

	// generate content
	// 	- package name
	// 	- imports
	// 	- tables
	// 	- conditions
	// 	- init function
	//
	content := p.generateContent(tables, nestedTables)
	{
		generatedFileName := p.getGeneratedFilePath(p.req.GetFileToGenerate()[0])
		p.res.File = append(p.res.File, &plugingo.CodeGeneratorResponse_File{
			Name:    proto.String(generatedFileName),
			Content: proto.String(content),
		})
	}

	// set supported features
	p.res.SupportedFeatures = proto.Uint64(uint64(plugingo.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))

	// format Go code and marshal protobuf
	//
	if err := goFmt(p.res); err != nil {
		log.Fatalf("Failed to format Go code: %v", err)
	}

	// marshal protobuf and write to stdout
	// 	- generated file name
	// 	- generated file content
	//
	data, err = proto.Marshal(p.res)
	if err != nil {
		log.Fatalf("Failed to marshal protobuf: %v", err)
	}

	// write to stdout
	if _, err := os.Stdout.Write(data); err != nil {
		log.Fatalf("Failed to write to stdout: %v", err)
	}
}

func (p *Plugin) EnableDefaultImports() {
	protoFile := getUserProtoFile(p.req)
	for _, m := range protoFile.GetMessageType() {
		for _, field := range m.GetField() {
			var typ = field.GetTypeName()
			switch *field.Type {
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				parts := strings.Split(typ, ".")
				typName := parts[len(parts)-1]
				if typName == "Timestamp" && parts[len(parts)-2] == "protobuf" && parts[len(parts)-3] == "google" {
					p.state.Imports.Enable(ImportTime)
				}
			}
		}
	}
}

// generateContent generates the string content for the output.
func (p *Plugin) generateContent(tables []Templater, nestedTables []Templater) string {
	builder := &strings.Builder{}

	builder.WriteString("// Code generated by protoc-gen-structify. DO NOT EDIT.\n")
	builder.WriteString("// source: ")
	builder.WriteString(p.req.GetFileToGenerate()[0] + "\n")
	builder.WriteString("// provider: ")
	builder.WriteString(p.provider.String() + "\n")

	conditions := p.BuildConditionsTemplate()
	inits := p.BuildInitFunctionTemplate()

	// write package name
	builder.WriteString("package " + p.PackageName + "\n\n")

	// write imports
	builder.WriteString(p.state.Imports.String())
	builder.WriteString(inits + "\n")

	// write nested tables
	for _, t := range nestedTables {
		builder.WriteString(t.BuildTemplate())
		builder.WriteString("\n")
	}

	// write tables
	for _, t := range tables {
		builder.WriteString(t.BuildTemplate())
		builder.WriteString("\n")
	}

	builder.WriteString(conditions + "\n")

	return builder.String()
}

// parseCommandLineParameters parses the command line parameters into the Param map.
func (p *Plugin) parseCommandLineParameters(parameter string) {
	p.Param = make(map[string]string)
	params := strings.Split(parameter, ",")
	for _, param := range params {
		if i := strings.Index(param, "="); i < 0 {
			p.Param[param] = ""
		} else {
			p.Param[param[:i]] = param[i+1:]
		}
	}
	p.parsePathType()
}

// parsePathType parses the path type from the parameters.
func (p *Plugin) parsePathType() {
	switch p.Param["paths"] {
	case "import":
		p.pathType = PathTypeImport
	case "source_relative":
		p.pathType = PathTypeSourceRelative
	default:
		log.Fatalf(`Unknown path type %q: want "import" or "source_relative".`, p.Param["paths"])
	}
}

// fillState fills the state with default values.
func (p *Plugin) fillState() *State {
	protoFile := getUserProtoFile(p.req)

	tableMessages, nestedTables := getMessages(p.req)
	return &State{
		Provider:     p.parseProvider(),
		PackageName:  protoFile.GetPackage(),
		FileName:     p.parseFileName(),
		Imports:      map[Import]bool{},
		Tables:       tableMessages,
		NestedTables: nestedTables,
	}
}

func (p *Plugin) parseFileName() string {
	fileBase := path.Base(p.req.GetFileToGenerate()[0])
	fileExt := path.Ext(fileBase)
	return strings.TrimSuffix(fileBase, fileExt)
}

// parsePackageName parses the package name from the protobuf options.
func (p *Plugin) parsePackageName() {
	for _, f := range getUserProtoFiles(p.req) {
		p.PackageName = f.GetPackage()
	}
}

// getGeneratedFilePath gets the generated file path based on the source file path.
func (p *Plugin) getGeneratedFilePath(sourceFilePath string) string {
	generatedBaseName := p.FileNameWithoutExt + GeneratedFilePostfix

	if p.pathType == PathTypeSourceRelative {
		// The generated file will have the same base as the source file, and it will be located in the same directory.
		fileDir := path.Dir(sourceFilePath)
		return path.Join(fileDir, generatedBaseName)
	}

	// If the path type is not source-relative, the generated file will have the same base as the source file,
	// but it will be located in the current directory.
	return generatedBaseName
}

// parseProvider parses the provider from the protobuf options.
func (p *Plugin) parseProvider() Provider {
	protoFile := getUserProtoFile(p.req)
	opts := getDBOptions(protoFile)
	if opts != nil {
		switch opts.GetProvider() {
		case "mysql":
			return ProviderMysql
		case "postgres":
			return ProviderPostgres
		case "sqlite":
			return ProviderSqlite
		default:
			return ProviderPostgres
		}
	}
	return ProviderPostgres
}

// getNestedTemplater parses the proto file and returns a slice of Tables.
func (p *Plugin) getNestedTemplater() ([]Templater, error) {
	var tables []Templater

	for _, m := range p.state.NestedTables {
		tables = append(tables, createNewNestedTemplate(m, p.state))
	}

	if len(tables) > 0 {
		p.state.Imports.Enable(ImportJson, ImportSQLDriver)
	}

	return tables, nil
}

// getTemplaterTables parses the proto file and returns a slice of Tables.
func (p *Plugin) getTemplaterTables() ([]Templater, error) {
	var tables []Templater

	for _, m := range p.state.Tables {
		var table Templater
		switch p.provider {
		case ProviderPostgres:
			p.state.Imports.Enable(ImportErrors, ImportContext)
			table = createNewPostgresTableTemplate(m, p.state)
		case ProviderMysql:
			// todo: implement
		default:
			return nil, ErrUnsupportedProvider
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// fillRelation fills the relations map in the state struct.
func (p *Plugin) fillRelation(state *State) *State {
	protoFile := getUserProtoFile(p.req)
	state.Relations = make(map[string]*Relation)

	for _, msg := range protoFile.GetMessageType() {
		for _, field := range msg.GetField() {
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				convertedType := convertType(field)

				relation := &Relation{
					Field:      detectField(detectStructName(convertedType)),
					Reference:  detectReference(msg.GetName()),
					TableName:  detectTableName(convertedType),  // Assuming msg.GetName() is the table name
					StructName: detectStructName(convertedType), // Assuming field.GetName() is the struct name
					Store:      detectStoreName(convertedType),  // Fill this with the proper value
					Many:       detectMany(convertedType),       // As the field is repeated, it means there are many relations
					Limit:      100,                             // default relation limit
				}

				options := getFieldOptions(field)
				if options != nil {
					relOptions := options.GetRelation()
					if relOptions != nil {
						relation.Field = relOptions.GetField()
						relation.Reference = relOptions.GetReference()
						relation.Limit = uint64(relOptions.GetLimit())
					}
				}

				if checkIsRelation(field) {
					// Add the relation to the map of relations
					state.Relations[msg.GetName()+"::"+relation.StructName] = relation
				}
			}
		}
	}

	for _, msg := range protoFile.GetMessageType() {
		for _, field := range msg.GetField() {
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				convertedType := convertType(field)
				structName := detectStructName(convertedType)
				if v, ok := state.Relations[structName+"::"+msg.GetName()]; ok {
					state.Relations[msg.GetName()+"::"+structName].Field = v.Reference
					state.Relations[msg.GetName()+"::"+structName].Reference = v.Field
				}
			}
		}
	}

	return state
}

// checkProtobufVersion checks that the protobuf version is supported.
func (p *Plugin) checkProtobufVersion() error {
	ver := p.req.GetCompilerVersion()

	// check protobuf version is supported (3.12.0 or later)
	if ver.GetMajor() < 3 || (ver.GetMajor() == 3 && ver.GetMinor() < 12) {
		return fmt.Errorf("unsupported protobuf version: %s, please upgrade to 3.12.0 or later", ver.String())
	}

	// check protobuf syntax is supported (proto3)
	if err := checkProtoSyntax(getUserProtoFile(p.req)); err != nil {
		return fmt.Errorf("unsupported protobuf syntax: %s, only 'proto3' is supported", getUserProtoFile(p.req).GetSyntax())
	}

	return nil
}

// fillNestedTableStructMapping fills the nested table struct mapping.
func (p *Plugin) fillNestedTableStructMapping(state *State) *State {
	f := getUserProtoFile(p.req)

	state.NestedTableStructMapping = make(map[string]NestedTableVal)
	for _, m := range f.GetMessageType() {
		for _, f := range m.GetField() {
			if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				convertedType := convertType(f)
				if len(m.GetNestedType()) > 0 {
					nests := descriptorToMap(m.GetNestedType())
					if _, ok := nests["JSON"+m.GetName()+detectStructName(convertedType)]; ok {
						state.NestedTableStructMapping[convertedType] = NestedTableVal{
							StructureName: "JSON" + m.GetName() + detectStructName(convertedType),
							HasType:       true,
						}
					}
				}

			}
		}
	}

	return state
}

// fillJSONTypes fills the json types.
// It is used to generate the json types.
func (p *Plugin) fillJSONTypes(state *State) *State {
	f := getUserProtoFile(p.req)
	state.JSONTypes = make(map[string]JSONType)
	for _, m := range f.GetMessageType() {
		for _, field := range m.GetField() {
			if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				convertedType := convertType(field)
				if isJSON(field, state) {
					// Note that the field type might not be a string in every case.
					// Consider using a switch statement on field.GetType() if you need more specific types.
					p.state.JSONTypes[m.GetName()+"::"+field.GetName()] = JSONType{
						StructureName: m.GetName(),
						FieldName:     field.GetName(),
						FieldType:     convertedType,
						TypeName:      buildJSONTypeName(m.GetName(), field.GetName()),
						//Template:      fmt.Sprintf(``+`type %s %s`+"\n", buildJSONTypeName(m.GetName(), field.GetName()), convertedType),
						Template: fmt.Sprintf(`type %s %s`+"\n", buildJSONTypeName(m.GetName(), field.GetName()), convertedType),
					}
				}
			}
		}
	}
	return state
}
