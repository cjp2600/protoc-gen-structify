package plugin

import (
	"bytes"
	"text/template"
)

const InitFunctionsTemplate = `
// {{ .Plugin.FileNameWithoutExt | upperClientName }} is a map of provider to init function.
type {{ .Plugin.FileNameWithoutExt | upperClientName }} struct {
	db *sql.DB
{{ range $key, $value := .Storages }}
{{ $key }} *{{ $value }}{{ end }}
}

// New{{ .Plugin.FileNameWithoutExt | upperClientName }} returns a new {{ .Plugin.FileNameWithoutExt | upperClientName }}. {{ .ExtraVar }}
func New{{ .Plugin.FileNameWithoutExt | upperClientName }}(db *sql.DB) *{{ .Plugin.FileNameWithoutExt | upperClientName }} {
	return &{{ .Plugin.FileNameWithoutExt | upperClientName }}{
		db: db,
{{ range $key, $value := .Storages }}
{{ $key }}: &{{ $value | sToLowerCamel }}{db: db},{{ end }}
	}
}

{{ range $value := .Messages }}
// {{ $value }} returns the {{ $value | sToLowerCamel }} store.
func (c *{{ $.Plugin.FileNameWithoutExt | upperClientName }}) {{ $value }}() *{{ $value | sToLowerCamel }}Store {
	return c.{{ $value | sToLowerCamel }}Store
}
{{ end }}

`

func (p *Plugin) BuildInitFunctionTemplate() string {
	type TemplateData struct {
		Plugin   *Plugin
		ExtraVar string
		Storages map[string]string
		Messages []string
	}

	data := TemplateData{
		Plugin:   p,
		ExtraVar: "extra value",
		Storages: make(map[string]string),
	}

	for _, m := range getMessages(p.req) {
		data.Messages = append(data.Messages, m.GetName())
		data.Storages[sToLowerCamel(m.GetName())+"Store"] = sToLowerCamel(m.GetName()) + "Store"
	}

	var output bytes.Buffer

	funcs := template.FuncMap{
		"upperClientName": upperClientName,
		"lowerClientName": lowerClientName,
		"sToCml":          sToCml,
		"sToLowerCamel":   sToLowerCamel,
	}

	tmpl, err := template.New("goFile").Funcs(funcs).Parse(InitFunctionsTemplate)
	if err != nil {
		return ""
	}

	if err = tmpl.Execute(&output, data); err != nil {
		return ""
	}

	// enable imports
	p.imports.Enable(ImportDb, ImportLibPQ)

	return output.String()
}
