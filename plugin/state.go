package plugin

import "google.golang.org/protobuf/types/descriptorpb"

// State is the state of the plugin.
type State struct {
	Provider    Provider // Provider is the provider of the plugin.
	PackageName string   // PackageName is the package name of the plugin.
	FileName    string   // FileName is the file name of the plugin.

	Imports      ImportSet
	Extensions   ExtensionSet
	Relations    map[string]*Relation
	Tables       []*descriptorpb.DescriptorProto
	NestedTables []*descriptorpb.DescriptorProto

	// NestedTableStructMapping is a map of nested table name to struct name.
	NestedTableStructMapping map[string]NestedTableVal

	JSONTypes map[string]JSONType
}

func (s *State) IsExistInNestedTables(name string) bool {
	for _, t := range s.NestedTables {
		if t.GetName() == name {
			return true
		}
	}
	return false
}

type JSONType struct {
	StructureName string
	FieldName     string
	FieldType     string
	TypeName      string
	Template      string
	Repeated      bool
}

type NestedTableVal struct {
	StructureName string
	HasType       bool
}

type Relation struct {
	Field      string
	Reference  string
	TableName  string
	StructName string
	Store      string
	Limit      uint64
	Many       bool
}

func (s *State) FillImports(tables []Templater) {
	for _, t := range tables {
		for i, v := range t.Imports() {
			if v {
				s.Imports[i] = v
			}
		}
	}
}
