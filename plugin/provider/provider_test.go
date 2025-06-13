package provider

import (
	"testing"

	structify "github.com/cjp2600/protoc-gen-structify/plugin/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	pluginpb "google.golang.org/protobuf/types/pluginpb"

	_ "github.com/cjp2600/protoc-gen-structify/plugin/provider/clickhouse"
	_ "github.com/cjp2600/protoc-gen-structify/plugin/provider/postgres"
	_ "github.com/cjp2600/protoc-gen-structify/plugin/provider/sqlite"
)

func TestParseFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Provider
	}{
		{
			name:     "mysql provider",
			input:    "mysql",
			expected: Mysql,
		},
		{
			name:     "postgres provider",
			input:    "postgres",
			expected: Postgres,
		},
		{
			name:     "sqlite provider",
			input:    "sqlite",
			expected: Sqlite,
		},
		{
			name:     "clickhouse provider",
			input:    "clickhouse",
			expected: Clickhouse,
		},
		{
			name:     "unknown provider defaults to postgres",
			input:    "unknown",
			expected: Postgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProvider_String(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected string
	}{
		{
			name:     "mysql string",
			provider: Mysql,
			expected: "mysql",
		},
		{
			name:     "postgres string",
			provider: Postgres,
			expected: "postgres",
		},
		{
			name:     "sqlite string",
			provider: Sqlite,
			expected: "sqlite",
		},
		{
			name:     "clickhouse string",
			provider: Clickhouse,
			expected: "clickhouse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTemplateBuilder(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		builder, err := GetTemplateBuilder(nil)
		require.Error(t, err)
		assert.Nil(t, builder)
	})

	t.Run("empty request", func(t *testing.T) {
		request := &pluginpb.CodeGeneratorRequest{}
		builder, err := GetTemplateBuilder(request)
		require.Error(t, err)
		assert.Nil(t, builder)
	})

	t.Run("valid request", func(t *testing.T) {
		fileOptions := &descriptorpb.FileOptions{}
		dbOpts := &structify.StructifyDBOptions{Provider: "postgres"}
		proto.SetExtension(fileOptions, structify.E_Db, dbOpts)

		request := &pluginpb.CodeGeneratorRequest{
			ProtoFile: []*descriptorpb.FileDescriptorProto{
				{
					Name:    proto.String("test.proto"),
					Package: proto.String("test"),
					Options: fileOptions,
				},
			},
			FileToGenerate: []string{"test.proto"},
		}

		builder, err := GetTemplateBuilder(request)
		require.NoError(t, err)
		assert.NotNil(t, builder)
	})
}
