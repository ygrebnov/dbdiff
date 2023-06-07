package dbdiff

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygrebnov/dbdiff/models"
)

var (
	mockIDField         = models.Field{Name: "mock_id_field", FieldType: "text", PrimaryKey: true, Attrs: "primary key"}
	mockTextField       = models.Field{Name: "mock_text_field", FieldType: "text", PrimaryKey: false, Attrs: "not null"}
	mockBooleanField    = models.Field{Name: "mock_boolean_field", FieldType: "boolean", PrimaryKey: false, Attrs: "not null"}
	mockTimestampField  = models.Field{Name: "mock_timestamp_field", FieldType: "timestamp", PrimaryKey: false, Attrs: "not null default current_timestamp"}
	mockFields          = []*models.Field{&mockIDField, &mockTextField, &mockBooleanField, &mockTimestampField}
	mockSqliteTable     = newMockTable("mock_sqlite_table", sqlite)
	mockPostgresqlTable = newMockTable("mock_postgresql_table", postgresql)
)

// newMockTable generates a mock table object.
func newMockTable(name string, dbType models.DatabaseType) models.Table {
	var schema string
	if dbType == sqlite {
		schema += fmt.Sprintf("CREATE TABLE %s (\n", name)
	}
	for _, f := range mockFields {
		schema += fmt.Sprintf("%s %s %s,\n", f.Name, strings.ToUpper(f.FieldType), strings.ToUpper(f.Attrs))
	}
	if dbType == sqlite {
		schema += "FOREIGN KEY(mock_parent_ref) REFERENCES mock_parent_table(mock_id_field) ON DELETE CASCADE"
	}

	fieldNameIndex := make(map[string]int, len(mockFields))
	for i, f := range mockFields {
		fieldNameIndex[f.Name] = i
	}

	return models.Table{
		Name:           name,
		Schema:         schema,
		PrimaryKey:     &mockIDField,
		Fields:         mockFields,
		FieldNameIndex: fieldNameIndex,
	}
}

func TestDatabaseTypeParse(t *testing.T) {
	var tests = []struct {
		table  models.Table
		dbType models.DatabaseType
	}{
		{mockSqliteTable, sqlite},
		{mockPostgresqlTable, postgresql},
	}

	for _, test := range tests {
		t.Run(test.table.Name, func(t *testing.T) {
			//given
			actualTable := models.Table{
				Name:           test.table.Name,
				Schema:         test.table.Schema,
				FieldNameIndex: make(map[string]int),
			}
			//when
			test.dbType.Parse(&actualTable)
			//then
			require.Equal(t, test.table, actualTable)
		})
	}
}
