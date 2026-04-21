package templater

import (
	"strings"
	"testing"

	importpkg "github.com/cjp2600/protoc-gen-structify/plugin/import"
	statepkg "github.com/cjp2600/protoc-gen-structify/plugin/state"
	"github.com/stretchr/testify/require"
)

func TestInitTemplate_DefaultConnections(t *testing.T) {
	s := &statepkg.State{
		Imports: importpkg.NewImportSet(),
	}

	tpl := NewInitTemplater(s)
	require.NotNil(t, tpl)

	out := tpl.BuildTemplate()
	require.NotEmpty(t, out)

	require.True(t, strings.Contains(out, "DBRead *sql.DB"))
	require.True(t, strings.Contains(out, "DBWrite *sql.DB"))
	require.True(t, strings.Contains(out, "config.DB.DBWrite = config.DB.DBRead"))
	require.True(t, strings.Contains(out, "func NewTxManager(db *sql.DB) *TxManager"))
	require.False(t, strings.Contains(out, "type DBReadConnection interface {"))
	require.False(t, strings.Contains(out, "dbWrite, ok := config.DB.DBRead.(DBWriteConnection)"))
}

func TestInitTemplate_SQLXConnections(t *testing.T) {
	s := &statepkg.State{
		Imports: importpkg.NewImportSet(),
		UseSQLX: true,
	}

	tpl := NewInitTemplater(s)
	require.NotNil(t, tpl)

	out := tpl.BuildTemplate()
	require.NotEmpty(t, out)

	require.True(t, strings.Contains(out, "type DBReadConnection interface {"))
	require.True(t, strings.Contains(out, "type DBWriteConnection interface {"))
	require.True(t, strings.Contains(out, "DBRead DBReadConnection"))
	require.True(t, strings.Contains(out, "DBWrite DBWriteConnection"))
	require.True(t, strings.Contains(out, "dbWrite, ok := config.DB.DBRead.(DBWriteConnection)"))
	require.True(t, strings.Contains(out, "func NewTxManager(db DBWriteConnection) *TxManager"))
	require.True(t, strings.Contains(out, "db QueryExecer"))
}
