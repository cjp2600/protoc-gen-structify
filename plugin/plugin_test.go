package plugin

import (
	"io"
	"os"
	"testing"

	options "github.com/cjp2600/protoc-gen-structify/plugin/options"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin()
	assert.NotNil(t, plugin)
	assert.NotNil(t, plugin.req)
	assert.NotNil(t, plugin.res)
}

func TestParseCommandLineParameters(t *testing.T) {
	tests := []struct {
		name     string
		params   string
		expected map[string]string
	}{
		{
			name:   "empty parameters",
			params: "paths=import", // Set default path type
			expected: map[string]string{
				"paths": "import",
			},
		},
		{
			name:   "single parameter without value",
			params: "include_connection,paths=import",
			expected: map[string]string{
				"include_connection": "",
				"paths":              "import",
			},
		},
		{
			name:   "single parameter with value",
			params: "paths=import",
			expected: map[string]string{
				"paths": "import",
			},
		},
		{
			name:   "multiple parameters",
			params: "paths=import,include_connection=true,create_crud_table_schemas=true",
			expected: map[string]string{
				"paths":                     "import",
				"include_connection":        "true",
				"create_crud_table_schemas": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewPlugin()
			plugin.parseCommandLineParameters(tt.params)
			assert.Equal(t, tt.expected, plugin.param)
		})
	}
}

func TestParseIncludeConnectionParam(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{
			name:     "true value",
			param:    "true",
			expected: true,
		},
		{
			name:     "false value",
			param:    "false",
			expected: false,
		},
		{
			name:     "empty value",
			param:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewPlugin()
			plugin.param = map[string]string{
				"include_connection": tt.param,
			}
			result := plugin.parseIncludeConnectionParam()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseCRUDSchemasParam(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{
			name:     "true value",
			param:    "true",
			expected: true,
		},
		{
			name:     "false value",
			param:    "false",
			expected: false,
		},
		{
			name:     "empty value",
			param:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewPlugin()
			plugin.param = map[string]string{
				"create_crud_table_schemas": tt.param,
			}
			result := plugin.parseCRUDSchemasParam()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileName(t *testing.T) {
	tests := []struct {
		name           string
		pathType       pathType
		fileToGenerate string
		inputName      string
		expected       string
	}{
		{
			name:           "import path type",
			pathType:       PathTypeImport,
			fileToGenerate: "test.proto",
			inputName:      "test",
			expected:       "test.db.go",
		},
		{
			name:           "source relative path type",
			pathType:       PathTypeSourceRelative,
			fileToGenerate: "path/to/test.proto",
			inputName:      "test",
			expected:       "path/to/test.db.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewPlugin()
			plugin.pathType = tt.pathType
			plugin.req.FileToGenerate = []string{tt.fileToGenerate}
			result := plugin.fileName(tt.inputName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckProtobufVersion(t *testing.T) {
	tests := []struct {
		name        string
		major       int32
		minor       int32
		patch       int32
		expectError bool
	}{
		{
			name:        "supported version",
			major:       3,
			minor:       12,
			patch:       0,
			expectError: false,
		},
		{
			name:        "unsupported version - too old",
			major:       3,
			minor:       11,
			patch:       0,
			expectError: true,
		},
		{
			name:        "unsupported version - major version too old",
			major:       2,
			minor:       0,
			patch:       0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewPlugin()
			plugin.req.CompilerVersion = &plugingo.Version{
				Major: proto.Int32(tt.major),
				Minor: proto.Int32(tt.minor),
				Patch: proto.Int32(tt.patch),
			}
			plugin.req.ProtoFile = []*descriptor.FileDescriptorProto{
				{
					Name:    proto.String("test.proto"),
					Package: proto.String("test"),
					Syntax:  proto.String("proto3"),
					MessageType: []*descriptor.DescriptorProto{
						{
							Name: proto.String("TestMessage"),
						},
					},
				},
			}
			plugin.req.FileToGenerate = []string{"test.proto"}

			err := plugin.checkProtobufVersion()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPluginRun(t *testing.T) {
	// Create a test request
	fileOptions := &descriptor.FileOptions{}
	proto.SetExtension(fileOptions, options.E_Db, &options.StructifyDBOptions{
		Provider: "postgres",
	})
	req := &plugingo.CodeGeneratorRequest{
		CompilerVersion: &plugingo.Version{
			Major: proto.Int32(3),
			Minor: proto.Int32(12),
			Patch: proto.Int32(0),
		},
		FileToGenerate: []string{"test.proto"},
		Parameter:      proto.String("paths=import"),
		ProtoFile: []*descriptor.FileDescriptorProto{
			{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				Syntax:  proto.String("proto3"),
				Options: fileOptions,
				MessageType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("TestMessage"),
					},
				},
			},
		},
	}

	// Marshal the request
	data, err := proto.Marshal(req)
	require.NoError(t, err)

	// Create a pipe to simulate stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r

	// Write the request to the pipe
	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)
		w.Close()
	}()

	// Create pipes for stdout
	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = stdoutW

	// Run the plugin
	plugin := NewPlugin()
	plugin.Run()

	// Close the write end of the pipe
	stdoutW.Close()

	// Read the response from the read end of the pipe
	responseData, err := io.ReadAll(stdoutR)
	require.NoError(t, err)

	// Read the response
	response := &plugingo.CodeGeneratorResponse{}
	err = proto.Unmarshal(responseData, response)
	require.NoError(t, err)

	// Verify the response
	assert.NotNil(t, response)
	assert.Equal(t, uint64(plugingo.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL), response.GetSupportedFeatures())
}
