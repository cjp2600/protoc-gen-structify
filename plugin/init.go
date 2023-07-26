package plugin

import (
	"bytes"
	"fmt"
	"text/template"
)

// InitFunctionsTemplate is the template for the init functions.
const InitFunctionsTemplate = `
// {{ .Plugin.FileNameWithoutExt | upperClientName }}Options are the options for the {{ .Plugin.FileNameWithoutExt | upperClientName }}.
type {{ .Plugin.FileNameWithoutExt | upperClientName }}Options struct {
    SSLMode string
    Timeout int
}

// Option is a function that configures the {{ .Plugin.FileNameWithoutExt | upperClientName }}.
type {{ .Plugin.FileNameWithoutExt | upperClientName }}Option func(*{{ .Plugin.FileNameWithoutExt | upperClientName }}Options)

// WithSSLMode sets the SSL mode for the database connection.
func WithSSLMode(sslMode string) {{ .Plugin.FileNameWithoutExt | upperClientName }}Option {
    return func(opts *{{ .Plugin.FileNameWithoutExt | upperClientName }}Options) {
        opts.SSLMode = sslMode
    }
}

// WithTimeout sets the timeout for the database connection.
func WithTimeout(timeout int) {{ .Plugin.FileNameWithoutExt | upperClientName }}Option {
    return func(opts *{{ .Plugin.FileNameWithoutExt | upperClientName }}Options) {
        opts.Timeout = timeout
    }
}

// DBConnect connects to the database and returns a *sql.DB.
func DBConnect(host string, port int, user string, password string, dbname string, opts ...{{ .Plugin.FileNameWithoutExt | upperClientName }}Option) (*sql.DB, error) {
    options := &{{ .Plugin.FileNameWithoutExt | upperClientName }}Options{}

    for _, opt := range opts {
        opt(options)
    }

    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s", host, port, user, password, dbname)

    if options.SSLMode != "" {
        psqlInfo += " sslmode=" + options.SSLMode
    }
    if options.Timeout != 0 {
        psqlInfo += " connect_timeout=" + strconv.Itoa(options.Timeout)
    }

    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err = db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}

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
{{ $key }}: &{{ $value | sToCml }}{db: db},{{ end }}
	}
}

{{ range $value := .Messages }}
// {{ $value }} returns the {{ $value | sToCml }} store.
func (c *{{ $.Plugin.FileNameWithoutExt | upperClientName }}) {{ $value }}() *{{ $value | sToCml }}Store {
	return c.{{ $value | sToLowerCamel }}Store
}
{{ end }}

func (c *{{ $.Plugin.FileNameWithoutExt | upperClientName }}) CreateTables() error {
	var err error
{{ range $value := .Messages }}
	_, err = c.db.Exec(c.{{ $value | sToLowerCamel }}Store.CreateTableSQL())
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
{{ end }}
	return nil
}

// error row not found
var ErrRowNotFound = errors.New("row not found")

// ErrNoTransaction is returned when the transaction is nil.
var ErrNoTransaction = errors.New("no transaction provided")

{{ range $key, $value := .JSONTypes }}
// {{ $key }} is a JSON type.
{{$value.Template}}

// New{{ $value.StructureName }}{{$value.FieldName | sToCml}} returns a new {{$value.TypeName}}.
func New{{ $value.StructureName }}{{ $value.FieldName | sToCml}}(val {{$value.FieldType}}) *{{$value.TypeName}} {
	value := {{$value.TypeName}}(val)
	return &value
}

// Scan implements the sql.Scanner interface for MyJSONType
func (m *{{$value.TypeName}}) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return fmt.Errorf("can't convert %T to JSONUserPhones", src)
}

// Value implements the driver.Valuer interface for {{$value.TypeName}}
func (m *{{$value.TypeName}}) Value() (driver.Value, error) {
	if m == nil {
		m = New{{ $value.StructureName }}{{ $value.FieldName | sToCml}}({{$value.FieldType}}{})
	}
	return json.Marshal(m)
}
{{ end }}
`

// BuildInitFunctionTemplate builds the init function template.
func (p *Plugin) BuildInitFunctionTemplate() string {
	type TemplateData struct {
		Plugin    *Plugin
		ExtraVar  string
		Storages  map[string]string
		Messages  []string
		JSONTypes map[string]*JSONType
	}

	// create the template data
	data := TemplateData{
		Plugin:    p,
		ExtraVar:  "extra value",
		Storages:  make(map[string]string),
		JSONTypes: make(map[string]*JSONType),
	}

	// add to state
	for _, jsonType := range p.state.JSONTypes {
		v := jsonType
		data.JSONTypes[jsonType.TypeName] = &v
	}

	if len(data.JSONTypes) > 0 {
		p.state.Imports.Enable(ImportJson, ImportSQLDriver)
	}

	if len(data.JSONTypes) > 0 {
		for k, jsonType := range data.JSONTypes {
			if jsonType.Repeated {
				jsonType.FieldType = "[]*" + k
				jsonType.TypeName = jsonType.TypeName + "Repeated"
			}
		}
	}

	// get the messages
	for _, m := range p.state.Tables {
		data.Messages = append(data.Messages, m.GetName())
		data.Storages[sToLowerCamel(m.GetName())+"Store"] = sToCml(m.GetName()) + "Store"
	}

	var output bytes.Buffer
	funcs := template.FuncMap{
		"upperClientName": upperClientName,
		"lowerClientName": lowerClientName,
		"sToCml":          sToCml,
		"sToLowerCamel":   sToLowerCamel,
	}

	// parse the template
	tmpl, err := template.New("goFile").Funcs(funcs).Parse(InitFunctionsTemplate)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// execute the template
	if err = tmpl.Execute(&output, data); err != nil {
		fmt.Println(err)
		return ""
	}

	// enable imports
	//
	p.state.Imports.Enable(ImportDb, ImportLibPQ, ImportStrconv)

	return output.String()
}
