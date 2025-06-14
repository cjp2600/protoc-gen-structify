package state

import (
	"testing"

	_import "github.com/cjp2600/protoc-gen-structify/plugin/import"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestNewState(t *testing.T) {
	// Create a test request
	req := &plugingo.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile: []*descriptor.FileDescriptorProto{
			{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				MessageType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("TestMessage"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name:   proto.String("id"),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
								Number: proto.Int32(1),
							},
						},
					},
				},
			},
		},
	}

	state := NewState(req)
	assert.NotNil(t, state)
	assert.Equal(t, "test", state.PackageName)
	assert.Equal(t, "test.proto", state.FileToGenerate)
	assert.NotNil(t, state.Imports)
	assert.NotNil(t, state.Messages)
	assert.NotNil(t, state.NestedMessages)
	assert.NotNil(t, state.Relations)
	assert.NotNil(t, state.SingleTypes)
}

func TestState_IsRelation(t *testing.T) {
	// Create a test request with a relation field
	req := &plugingo.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile: []*descriptor.FileDescriptorProto{
			{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				MessageType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("User"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name:   proto.String("id"),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
								Number: proto.Int32(1),
							},
							{
								Name:     proto.String("posts"),
								Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
								TypeName: proto.String(".test.Post"),
								Number:   proto.Int32(2),
								Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
							},
						},
					},
					{
						Name: proto.String("Post"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name:     proto.String("user"),
								Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
								TypeName: proto.String(".test.User"),
								Number:   proto.Int32(2),
							},
						},
					},
				},
			},
		},
	}

	state := NewState(req)
	require.NotNil(t, state)

	// Test IsRelation for a relation field
	userMsg := state.Messages.FindByName("User")
	require.NotNil(t, userMsg)
	postsField := userMsg.GetField()[1]
	assert.True(t, state.IsRelation(postsField))

	// Test IsRelation for a non-relation field
	idField := userMsg.GetField()[0]
	assert.False(t, state.IsRelation(idField))
}

func TestState_IsExistInTables(t *testing.T) {
	req := &plugingo.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile: []*descriptor.FileDescriptorProto{
			{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				MessageType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("TestMessage"),
					},
				},
			},
		},
	}

	state := NewState(req)
	require.NotNil(t, state)

	assert.True(t, state.IsExistInTables("TestMessage"))
	assert.False(t, state.IsExistInTables("NonExistentMessage"))
}

func TestState_IsExistInNestedTables(t *testing.T) {
	req := &plugingo.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile: []*descriptor.FileDescriptorProto{
			{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				MessageType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("OuterMessage"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name:     proto.String("nested"),
								Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
								TypeName: proto.String(".test.OuterMessage.NestedMessage"),
								Number:   proto.Int32(1),
							},
						},
						NestedType: []*descriptor.DescriptorProto{
							{
								Name: proto.String("NestedMessage"),
							},
						},
					},
				},
			},
		},
	}

	state := NewState(req)
	require.NotNil(t, state)

	assert.True(t, state.IsExistInNestedTables("NestedMessage"))
	assert.False(t, state.IsExistInNestedTables("NonExistentNestedMessage"))
}

func TestState_ImportsFromTable(t *testing.T) {
	// Create a mock templater
	mockTemplater := &mockTemplater{
		imports: _import.NewImportSet(),
	}
	mockTemplater.imports.Enable(
		_import.ImportDb,
		_import.ImportFMT,
	)

	state := &State{
		Imports: _import.NewImportSet(),
	}

	state.ImportsFromTable([]Templater{mockTemplater})

	// Check if imports were added
	imports := state.Imports.GetImports()
	foundDb := false
	foundFmt := false
	for _, imp := range imports {
		if imp == _import.ImportDb {
			foundDb = true
		}
		if imp == _import.ImportFMT {
			foundFmt = true
		}
	}
	assert.True(t, foundDb, "ImportDb should be present")
	assert.True(t, foundFmt, "ImportFMT should be present")
}

// mockTemplater implements Templater interface for testing
type mockTemplater struct {
	imports *_import.ImportSet
}

func (m *mockTemplater) TemplateName() string {
	return "mock"
}

func (m *mockTemplater) BuildTemplate() string {
	return "mock template"
}

func (m *mockTemplater) Imports() *_import.ImportSet {
	return m.imports
}

func TestFlattenMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  *descriptor.DescriptorProto
		parent   string
		expected map[string]*MessageDescriptor
	}{
		{
			name: "simple nested message",
			message: &descriptor.DescriptorProto{
				Name: proto.String("User"),
				NestedType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("Profile"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name: proto.String("name"),
								Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
						},
					},
				},
			},
			parent: "",
			expected: map[string]*MessageDescriptor{
				"Profile": {
					Descriptor: &descriptor.DescriptorProto{
						Name: proto.String("Profile"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name: proto.String("name"),
								Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
						},
					},
					StructureName: "UserProfile",
					SourceName:    "Profile",
				},
			},
		},
		{
			name: "nested message with same name as parent",
			message: &descriptor.DescriptorProto{
				Name: proto.String("User"),
				NestedType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("User"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name: proto.String("name"),
								Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
						},
					},
				},
			},
			parent: "",
			expected: map[string]*MessageDescriptor{
				"User": {
					Descriptor: &descriptor.DescriptorProto{
						Name: proto.String("User"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name: proto.String("name"),
								Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
						},
					},
					StructureName: "UserUser",
					SourceName:    "User",
				},
			},
		},
		{
			name: "deeply nested messages",
			message: &descriptor.DescriptorProto{
				Name: proto.String("User"),
				NestedType: []*descriptor.DescriptorProto{
					{
						Name: proto.String("Profile"),
						NestedType: []*descriptor.DescriptorProto{
							{
								Name: proto.String("Settings"),
								Field: []*descriptor.FieldDescriptorProto{
									{
										Name: proto.String("theme"),
										Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
									},
								},
							},
						},
					},
				},
			},
			parent: "",
			expected: map[string]*MessageDescriptor{
				"Profile": {
					Descriptor: &descriptor.DescriptorProto{
						Name: proto.String("Profile"),
						NestedType: []*descriptor.DescriptorProto{
							{
								Name: proto.String("Settings"),
								Field: []*descriptor.FieldDescriptorProto{
									{
										Name: proto.String("theme"),
										Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
									},
								},
							},
						},
					},
					StructureName: "UserProfile",
					SourceName:    "Profile",
				},
				"Settings": {
					Descriptor: &descriptor.DescriptorProto{
						Name: proto.String("Settings"),
						Field: []*descriptor.FieldDescriptorProto{
							{
								Name: proto.String("theme"),
								Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
						},
					},
					StructureName: "UserProfileSettings",
					SourceName:    "Settings",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]*MessageDescriptor)
			flattenMessage(tt.message, result, tt.parent)

			// Check that all expected messages are present
			for expectedName, expectedDesc := range tt.expected {
				actualDesc, exists := result[expectedName]
				assert.True(t, exists, "Expected message %s not found", expectedName)
				assert.Equal(t, expectedDesc.StructureName, actualDesc.StructureName,
					"Structure name mismatch for %s", expectedName)
				assert.Equal(t, expectedDesc.SourceName, actualDesc.SourceName,
					"Source name mismatch for %s", expectedName)
			}

			// Check that there are no unexpected messages
			assert.Equal(t, len(tt.expected), len(result),
				"Unexpected number of messages in result")
		})
	}
}

func TestGetLinkedStructName(t *testing.T) {
	tests := []struct {
		name          string
		message       *descriptor.DescriptorProto
		field         *descriptor.FieldDescriptorProto
		messages      NestedMessages
		convertedType string
		expected      string
	}{
		{
			name: "simple nested message",
			message: &descriptor.DescriptorProto{
				Name: proto.String("User"),
			},
			field: &descriptor.FieldDescriptorProto{
				Name:     proto.String("profile"),
				TypeName: proto.String(".User.Profile"),
			},
			messages: NestedMessages{
				"User.Profile": {
					Descriptor: &descriptor.DescriptorProto{
						Name: proto.String("Profile"),
					},
					StructureName: "UserProfile",
					SourceName:    "Profile",
				},
			},
			convertedType: "User.Profile",
			expected:      "UserProfile",
		},
		{
			name: "nested message with same name as parent",
			message: &descriptor.DescriptorProto{
				Name: proto.String("User"),
			},
			field: &descriptor.FieldDescriptorProto{
				Name:     proto.String("user"),
				TypeName: proto.String(".User.User"),
			},
			messages: NestedMessages{
				"User.User": {
					Descriptor: &descriptor.DescriptorProto{
						Name: proto.String("User"),
					},
					StructureName: "UserUser",
					SourceName:    "User",
				},
			},
			convertedType: "User.User",
			expected:      "UserUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLinkedStructName(tt.message, tt.field, tt.messages, tt.convertedType)
			assert.Equal(t, tt.expected, result, "Linked struct name mismatch")
		})
	}
}
