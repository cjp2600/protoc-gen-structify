package provider

import (
	"errors"

	helperpkg "github.com/cjp2600/structify/plugin/pkg/helper"
	"github.com/cjp2600/structify/plugin/provider/postgres"
	statepkg "github.com/cjp2600/structify/plugin/state"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
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
		case Mysql:
			// Return MySQL provider
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
	default:
		return Postgres
	}
}

// String returns the provider as a string.
func (p Provider) String() string {
	return string(p)
}

// Available providers.
var (
	Mysql    Provider = "mysql"
	Postgres Provider = "postgres"
	Sqlite   Provider = "sqlite"
)

// ErrUnsupportedProvider is returned when the provider is not supported.
var ErrUnsupportedProvider = errors.New("unsupported provider")
