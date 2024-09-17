package table

import (
	databaseSQL "database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygrebnov/dbdiff/entity/field"
	"github.com/ygrebnov/dbdiff/sql"
)

var (
	mockIDField = field.Field{
		Name:         "mock_id_field",
		FieldType:    "text",
		IsPrimaryKey: true,
		Attrs:        "primary key",
	}
	mockTextField = field.Field{
		Name:         "mock_text_field",
		FieldType:    "text",
		IsPrimaryKey: false,
		Attrs:        "not null",
	}
	mockBooleanField = field.Field{
		Name:         "mock_boolean_field",
		FieldType:    "boolean",
		IsPrimaryKey: false,
		Attrs:        "not null",
	}
	mockTimestampField = field.Field{
		Name:         "mock_timestamp_field",
		FieldType:    "timestamp",
		IsPrimaryKey: false,
		Attrs:        "not null default current_timestamp",
	}
	mockFields = []*field.Field{&mockIDField, &mockTextField, &mockBooleanField, &mockTimestampField}
)

func newMockSchema(name string) string {
	var schema string
	if name == "sqlite" {
		schema += fmt.Sprintf("CREATE TABLE %s (\n", name)
	}
	for _, f := range mockFields {
		schema += fmt.Sprintf("%s %s %s,\n", f.Name, strings.ToUpper(f.FieldType), strings.ToUpper(f.Attrs))
	}
	if name == "sqlite" {
		schema += "FOREIGN KEY(mock_parent_ref) REFERENCES mock_parent_table(mock_id_field) ON DELETE CASCADE"
	}

	return schema
}

// newMockTable generates a mock Table object.
func newMockTable(name, schema string) table {
	fieldsIndexes := make(map[string]int, len(mockFields))
	for i, f := range mockFields {
		fieldsIndexes[f.Name] = i
	}

	return table{
		name:          name,
		schema:        schema,
		primaryKey:    &mockIDField,
		fields:        mockFields,
		fieldsIndexes: fieldsIndexes,
	}
}

type mockRows struct {
	name, schema databaseSQL.NullString
	iterated     bool // will respond true only on one Next() call.
}

func (m *mockRows) Next() bool {
	if !m.iterated {
		m.iterated = true
		return true
	}

	return false
}

func (m *mockRows) Scan(dest ...interface{}) error {
	if len(dest) == 1 {
		schema := dest[0].(*databaseSQL.NullString)
		*schema = m.schema
	} else {
		name := dest[0].(*databaseSQL.NullString)
		*name = m.name

		schema := dest[1].(*databaseSQL.NullString)
		*schema = m.schema
	}

	return nil
}

func newMockRows(name, schema string) sql.Rows {
	return &mockRows{
		name:   databaseSQL.NullString{String: name, Valid: true},
		schema: databaseSQL.NullString{String: schema, Valid: true},
	}
}

func TestNew(t *testing.T) {
	var tests = []struct {
		name   string
		parser Parser
	}{
		{"sqlite", NewSqliteParser()},
		{"postgresql", NewPostgresqlParser()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := newMockSchema(test.name)

			tbl, err := New(newMockRows(test.name, schema), test.parser)
			require.NoError(t, err)
			require.NotNil(t, tbl)

			require.Equal(t, schema, tbl.GetSchema())
			require.ElementsMatch(t, mockFields, tbl.GetFields())

			fieldNames := []string{mockBooleanField.Name, mockIDField.Name, mockTextField.Name, mockTimestampField.Name}
			require.Equal(t, fieldNames, tbl.GetFieldNames())
		})
	}
}
