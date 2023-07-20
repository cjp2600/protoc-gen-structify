package plugin

// State is the state of the plugin.
type State struct {
	Provider    Provider // Provider is the provider of the plugin.
	PackageName string   // PackageName is the package name of the plugin.
	FileName    string   // FileName is the file name of the plugin.

	Imports    ImportSet
	Extensions ExtensionSet
	Relations  map[string]*Relation
}

type Relation struct {
	Field      string
	Reference  string
	TableName  string
	StructName string
	Store      string
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
