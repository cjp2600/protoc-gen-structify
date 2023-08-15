package postgres

import (
	importpkg "github.com/cjp2600/structify/plugin/import"
	templaterpkg "github.com/cjp2600/structify/plugin/provider/postgres/templater"
	statepkg "github.com/cjp2600/structify/plugin/state"
)

// Postgres is a type for providing content.
type Postgres struct{}

// GetInitStatement returns the initialization statement.
func (p *Postgres) GetInitStatement(s *statepkg.State) (statepkg.Templater, error) {
	templater := templaterpkg.NewInitTemplater(s)
	s.ImportsFromTable([]statepkg.Templater{templater})

	return templater, nil
}

// GetEntities returns the tables.
func (p *Postgres) GetEntities(state *statepkg.State) ([]statepkg.Templater, error) {
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
func (p *Postgres) GetFinalizeStatement(s *statepkg.State) (statepkg.Templater, error) {
	var table statepkg.Templater
	//table = NewInitStatement(s)

	s.ImportsFromTable([]statepkg.Templater{table})
	return table, nil
}
