package sqlite

import (
	importpkg "github.com/cjp2600/protoc-gen-structify/plugin/import"
	templaterpkg "github.com/cjp2600/protoc-gen-structify/plugin/provider/sqlite/templater"
	statepkg "github.com/cjp2600/protoc-gen-structify/plugin/state"
)

type Sqlite struct{}

func (s Sqlite) GetInitStatement(state *statepkg.State) (statepkg.Templater, error) {
	templater := templaterpkg.NewInitTemplater(state)
	state.ImportsFromTable([]statepkg.Templater{templater})

	return templater, nil
}

func (s Sqlite) GetEntities(state *statepkg.State) ([]statepkg.Templater, error) {
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

func (s Sqlite) GetFinalizeStatement(state *statepkg.State) (statepkg.Templater, error) {
	var table statepkg.Templater
	//table = NewInitStatement(s)

	state.ImportsFromTable([]statepkg.Templater{table})
	return table, nil
}
