package sqlite

import (
	"testing"

	importpkg "github.com/cjp2600/protoc-gen-structify/plugin/import"
	"github.com/cjp2600/protoc-gen-structify/plugin/state"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlite_GetInitStatement(t *testing.T) {
	p := &Sqlite{}
	s := &state.State{
		Imports: importpkg.NewImportSet(),
	}

	templater, err := p.GetInitStatement(s)
	require.NoError(t, err)
	assert.NotNil(t, templater)
}

func TestSqlite_GetEntities(t *testing.T) {
	tests := []struct {
		name          string
		state         *state.State
		expectedCount int
		expectError   bool
	}{
		{
			name: "empty state",
			state: &state.State{
				Imports: importpkg.NewImportSet(),
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "state with messages",
			state: &state.State{
				Imports: importpkg.NewImportSet(),
				Messages: []*descriptor.DescriptorProto{
					{
						Name: proto.String("TestMessage"),
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Sqlite{}
			templaters, err := p.GetEntities(tt.state)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, templaters)
			} else {
				require.NoError(t, err)
				assert.Len(t, templaters, tt.expectedCount)
			}
		})
	}
}

func TestSqlite_GetFinalizeStatement(t *testing.T) {
	p := &Sqlite{}
	s := &state.State{
		Imports: importpkg.NewImportSet(),
	}

	templater, err := p.GetFinalizeStatement(s)
	require.NoError(t, err)
	assert.Nil(t, templater) // Currently returns nil as per implementation
}
