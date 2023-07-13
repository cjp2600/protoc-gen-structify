package plugin

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// StructifyPlugin implements the protoc-gen-gogo plugin interface
// and generates structify code for the given file.
type StructifyPlugin struct {
	*generator.Generator
}

func NewStructifyPlugin() *StructifyPlugin {
	return &StructifyPlugin{}
}

// Name identifies the plugin
func (s StructifyPlugin) Name() string {
	return pluginName
}

func (s StructifyPlugin) Init(g *generator.Generator) {
	//TODO implement me
	panic("implement me")
}

func (s StructifyPlugin) Generate(file *generator.FileDescriptor) {
	//TODO implement me
	panic("implement me")
}

func (s StructifyPlugin) GenerateImports(file *generator.FileDescriptor) {
	//TODO implement me
	panic("implement me")
}

const pluginName = "structify"
