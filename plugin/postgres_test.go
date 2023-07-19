package plugin

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateCreateSQL(t *testing.T) {
	testCases := []struct {
		desc     string
		table    *PostgresTable
		expected string
	}{
		{
			desc: "Primary key, unique, not null",
			table: &PostgresTable{
				Name:      "TestTable",
				TableName: "test_table",
				Fields: []*Field{
					{
						Name:       "ID",
						SourceName: "id",
						DBType:     "TEXT",
						Options: Options{
							PrimaryKey: true,
							Unique:     true,
							Nullable:   false,
						},
					},
				},
			},
			expected: `CREATE TABLE IF NOT EXISTS test_table (
id TEXT PRIMARY KEY UNIQUE NOT NULL);`,
		},
		{
			desc: "AutoIncrement, default",
			table: &PostgresTable{
				Name:      "AnotherTable",
				TableName: "another_table",
				Fields: []*Field{
					{
						Name:       "Counter",
						SourceName: "counter",
						DBType:     "BIGINT",
						Options: Options{
							AutoIncrement: true,
							Default:       "0",
						},
					},
				},
			},
			expected: `CREATE TABLE IF NOT EXISTS another_table (
counter BIGINT SERIAL NOT NULL DEFAULT 0);`,
		},
		{
			desc: "Nullable field",
			table: &PostgresTable{
				Name:      "NullableTable",
				TableName: "nullable_table",
				Fields: []*Field{
					{
						Name:       "OptionalField",
						SourceName: "optional_field",
						DBType:     "TEXT",
						Options: Options{
							Nullable: true,
						},
					},
				},
			},
			expected: `CREATE TABLE IF NOT EXISTS nullable_table (
optional_field TEXT);`,
		},
		{
			desc: "Default UUID",
			table: &PostgresTable{
				Name:      "UUIDTable",
				TableName: "uuid_table",
				Fields: []*Field{
					{
						Name:       "UUIDField",
						SourceName: "uuid_field",
						DBType:     "UUID",
						Options: Options{
							Default: "uuid_generate_v4()",
						},
					},
				},
			},
			expected: `CREATE TABLE IF NOT EXISTS uuid_table (
uuid_field UUID NOT NULL DEFAULT uuid_generate_v4());`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			sql := tc.table.GenerateCreateSQL()
			require.Equal(t, tc.expected, sql)
		})
	}
}
