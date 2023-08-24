package plugin

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"

	generatorpkg "github.com/cjp2600/protoc-gen-structify/plugin/generator"
	helperpkg "github.com/cjp2600/protoc-gen-structify/plugin/pkg/helper"
	"github.com/cjp2600/protoc-gen-structify/plugin/provider"
	statepkg "github.com/cjp2600/protoc-gen-structify/plugin/state"
)

// Plugin handles generation of code based on protobuf.
type Plugin struct {
	req      *plugingo.CodeGeneratorRequest
	res      *plugingo.CodeGeneratorResponse
	state    *statepkg.State
	pathType pathType
	param    map[string]string
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

	// check if there are any proto files
	if len(helperpkg.GetUserProtoFiles(p.req)) == 0 {
		log.Fatalf("No proto file is supported: %d", len(helperpkg.GetUserProtoFiles(p.req)))
	}

	// only one proto file is supported
	if len(helperpkg.GetUserProtoFiles(p.req)) > 1 {
		log.Fatalf("Only one proto file is supported: %d", len(helperpkg.GetUserProtoFiles(p.req)))
	}

	// parse command line parameters
	{
		p.parseCommandLineParameters(p.req.GetParameter())
	}

	// get default plugin state
	p.state = statepkg.NewState(p.req)
	{
		// set additional state parameters
		p.state.IncludeConnection = p.parseIncludeConnectionParam()
	}

	// get provider template builder based on command line parameter
	templBuilder, err := provider.GetTemplateBuilder(p.req)
	if err != nil {
		log.Fatalf("Failed to get template builder: %v", err)
	}

	// generate main content
	// 	- package name
	// 	- imports
	// 	- init block
	// 	- messages block
	// 	- conditions block
	//
	generator := generatorpkg.NewContentGenerator(p.state, templBuilder)
	content, err := generator.Content()
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	p.res.File = append(p.res.File, &plugingo.CodeGeneratorResponse_File{
		Name:    proto.String(p.getGeneratedFilePath(p.req.GetFileToGenerate()[0])),
		Content: proto.String(content),
	})

	// set supported features
	p.res.SupportedFeatures = proto.Uint64(uint64(plugingo.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))

	// format Go code and marshal protobuf
	//
	if err := helperpkg.GoFmt(p.res); err != nil {
		log.Fatalf("Failed to format Go code: %v", err)
	}

	// marshal protobuf and write to stdout
	data, err = proto.Marshal(p.res)
	if err != nil {
		log.Fatalf("Failed to marshal protobuf: %v", err)
	}

	// write to stdout
	if _, err := os.Stdout.Write(data); err != nil {
		log.Fatalf("Failed to write to stdout: %v", err)
	}
}

// parseCommandLineParameters parses the command line parameters into the param map.
func (p *Plugin) parseCommandLineParameters(parameter string) {
	p.param = make(map[string]string)
	params := strings.Split(parameter, ",")
	for _, param := range params {
		if i := strings.Index(param, "="); i < 0 {
			p.param[param] = ""
		} else {
			p.param[param[:i]] = param[i+1:]
		}
	}
	p.parsePathType()
}

func (p *Plugin) parseIncludeConnectionParam() bool {
	return p.param["include_connection"] == "true"
}

// parsePathType parses the path type from the parameters.
func (p *Plugin) parsePathType() {
	switch p.param["paths"] {
	case "import":
		p.pathType = PathTypeImport
	case "source_relative":
		p.pathType = PathTypeSourceRelative
	default:
		log.Fatalf(`Unknown path type %q: want "import" or "source_relative".`, p.param["paths"])
	}
}

// getGeneratedFilePath gets the generated file path based on the source file path.
func (p *Plugin) getGeneratedFilePath(sourceFilePath string) string {
	generatedBaseName := p.state.FileName + GeneratedFilePostfix

	if p.pathType == PathTypeSourceRelative {
		// The generated file will have the same base as the source file, and it will be located in the same directory.
		fileDir := path.Dir(sourceFilePath)
		return path.Join(fileDir, generatedBaseName)
	}

	// If the path type is not source-relative, the generated file will have the same base as the source file,
	// but it will be located in the current directory.
	return generatedBaseName
}

// checkProtobufVersion checks that the protobuf version is supported.
func (p *Plugin) checkProtobufVersion() error {
	ver := p.req.GetCompilerVersion()

	// check protobuf version is supported (3.12.0 or later)
	if ver.GetMajor() < 3 || (ver.GetMajor() == 3 && ver.GetMinor() < 12) {
		return fmt.Errorf("unsupported protobuf version: %s, please upgrade to 3.12.0 or later", ver.String())
	}

	// check protobuf syntax is supported (proto3)
	if err := helperpkg.CheckProtoSyntax(helperpkg.GetUserProtoFile(p.req)); err != nil {
		return fmt.Errorf("unsupported protobuf syntax: %s, only 'proto3' is supported", helperpkg.GetUserProtoFile(p.req).GetSyntax())
	}

	return nil
}

// pathType is a type for how to generate output filenames.
type pathType int

// Available path singleTypes.
const (
	PathTypeImport pathType = iota
	PathTypeSourceRelative
)

const GeneratedFilePostfix = ".db.go"
