package clickhouse

import (
	importpkg "github.com/cjp2600/protoc-gen-structify/plugin/import"
	templaterpkg "github.com/cjp2600/protoc-gen-structify/plugin/provider/postgres/templater"
	statepkg "github.com/cjp2600/protoc-gen-structify/plugin/state"
)

// Clickhouse is a type for providing content.
type Clickhouse struct{}

// GetInitStatement returns the initialization statement.
func (p *Clickhouse) GetInitStatement(s *statepkg.State) (statepkg.Templater, error) {
	templater := templaterpkg.NewInitTemplater(s)
	s.ImportsFromTable([]statepkg.Templater{templater})

	return templater, nil
}

// GetEntities returns the tables.
func (p *Clickhouse) GetEntities(state *statepkg.State) ([]statepkg.Templater, error) {
	var models []statepkg.Templater

	// set usage imports
	state.Imports.Enable(
		importpkg.ImportErrors,
		importpkg.ImportContext,
	)

	for _, message := range state.Messages {
		models = append(models, templaterpkg.NewTableTemplater(message, state))
	}

	state.ImportsFromTable(models)
	return models, nil
}

// GetFinalizeStatement returns the finalization statement.
func (p *Clickhouse) GetFinalizeStatement(s *statepkg.State) (statepkg.Templater, error) {
	var table statepkg.Templater
	//table = NewInitStatement(s)

	s.ImportsFromTable([]statepkg.Templater{table})
	return table, nil
}
