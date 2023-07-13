package plugin

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// StructifyPlugin implements the protoc-gen-gogo plugin interface
// and generates structify code for the given file.
type StructifyPlugin struct {
	*generator.Generator
	generator.PluginImports
	EmptyFiles     []string
	currentPackage string
	currentFile    *generator.FileDescriptor
	generateCrud   bool

	PrivateEntities map[string]PrivateEntity
	Fields          map[string][]*descriptor.FieldDescriptorProto
}

func newStructifyPlugin(gen *generator.Generator) *StructifyPlugin {
	return &StructifyPlugin{Generator: gen}
}

// Name identifies the plugin
func (s *StructifyPlugin) Name() string {
	return pluginName
}

// Init initializes the plugin
func (s *StructifyPlugin) Init(g *generator.Generator) {
	// register the plugin with the generator
	generator.RegisterPlugin(newStructifyPlugin(g))

	// set the generator so we can use it later
	s.Generator = g
}

func (s *StructifyPlugin) Generate(file *generator.FileDescriptor) {
	s.P(`var a string = "hello"`)
}

func (s *StructifyPlugin) GenerateImports(file *generator.FileDescriptor) {
	//s.P(`var b string = "world"`)
}

const pluginName = "structify"

type PrivateEntity struct {
	name    string
	items   []*descriptor.FieldDescriptorProto
	message *generator.Descriptor
}
