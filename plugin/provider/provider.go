package provider

import (
	"errors"
	"github.com/cjp2600/protoc-gen-structify/plugin/provider/clickhouse"

	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"

	helperpkg "github.com/cjp2600/protoc-gen-structify/plugin/pkg/helper"
	"github.com/cjp2600/protoc-gen-structify/plugin/provider/postgres"
	"github.com/cjp2600/protoc-gen-structify/plugin/provider/sqlite"
	statepkg "github.com/cjp2600/protoc-gen-structify/plugin/state"
)

// TemplateBuilder is a type for providing content.
// It is used to build the template.
type TemplateBuilder interface {
	// GetInitStatement returns the initialization statement.
	GetInitStatement(*statepkg.State) (statepkg.Templater, error)
	// GetEntities returns the entities.
	GetEntities(*statepkg.State) ([]statepkg.Templater, error)
	// GetFinalizeStatement returns the finalization statement.
	GetFinalizeStatement(*statepkg.State) (statepkg.Templater, error)
}

// GetTemplateBuilder returns the TemplateBuilder for the given provider.
func GetTemplateBuilder(request *plugingo.CodeGeneratorRequest) (TemplateBuilder, error) {
	protoFile := helperpkg.GetUserProtoFile(request)
	opts := helperpkg.GetDBOptions(protoFile)
	if opts != nil {
		switch ParseFromString(opts.GetProvider()) {
		case Postgres:
			return &postgres.Postgres{}, nil
		case Sqlite:
			return &sqlite.Sqlite{}, nil
		case Clickhouse:
			return &clickhouse.Clickhouse{}, nil
		}
	}
	return nil, ErrUnsupportedProvider
}

// Provider represents the database provider.
type Provider string

// ParseFromString parses the Provider from the protobuf options.
func ParseFromString(provider string) Provider {
	switch provider {
	case "mysql":
		return Mysql
	case "postgres":
		return Postgres
	case "sqlite":
		return Sqlite
	case "clickhouse":
		return Clickhouse
	default:
		// Default to Postgres
		return Postgres
	}
}

// String returns the provider as a string.
func (p Provider) String() string {
	return string(p)
}

// Available providers.
var (
	Mysql      Provider = "mysql"
	Postgres   Provider = "postgres"
	Sqlite     Provider = "sqlite"
	Clickhouse Provider = "clickhouse"
)

// ErrUnsupportedProvider is returned when the provider is not supported.
var ErrUnsupportedProvider = errors.New("unsupported provider")
