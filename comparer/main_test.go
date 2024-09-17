package comparer

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/ygrebnov/dbdiff/entity/field"
	"github.com/ygrebnov/dbdiff/entity/row"
	"github.com/ygrebnov/dbdiff/entity/table"
)

type mockRow struct {
	values []string
	pk     string
}

func (m *mockRow) GetPKValue() string {
	return m.pk
}

func (m *mockRow) GetValues() []string {
	return m.values
}

func (m *mockRow) Parse() error {
	return nil
}

func newMockRow(values ...string) row.Row {
	return &mockRow{values: values, pk: "pk"}
}

type mockTable struct {
	table.Table
	id         uuid.UUID
	fieldNames []string
}

func (m *mockTable) GetID() uuid.UUID {
	return m.id
}

func (m *mockTable) GetName() string {
	return "table1"
}

func (m *mockTable) GetPrimaryKey() *field.Field {
	return &field.Field{Name: "pk"}
}

func (m *mockTable) GetFieldNames() []string {
	return m.fieldNames
}

func newMockTable(id uuid.UUID, fieldNames []string) table.Table {
	return &mockTable{id: id, fieldNames: fieldNames}
}

var (
	s1, s2, v1, v2, v3    = "s1", "s2", "v1", "v2", "v3"
	mockTable1Column      = newMockTable(uuid.New(), []string{v1})
	mockTable2Columns     = newMockTable(uuid.New(), []string{v1, v2})
	mockTable3Columns     = newMockTable(uuid.New(), []string{v1, v2, v3})
	differencesLeftEmpty  = Differences{{v1, "", s2}, {v2, "", s2}}
	differencesRightEmpty = Differences{{v1, s1, ""}, {v2, s1, ""}}
	differencesEqual      = Differences{{v1, s1, s1}, {v2, s2, s2}}
	differencesDifferent  = Differences{{v1, s1, s2}, {v2, s2, s1}}
)

func mockOutput(diffs Differences, verbose bool) string {
	output := "    Field   Database1   Database2\n"

	for _, d := range diffs {
		var prefix string

		switch {
		case !verbose:
			prefix = strings.Repeat(" ", 4)
		case d.Value1 == d.Value2:
			prefix = "  = "
		case d.Value1 != d.Value2:
			prefix = "  x "
		}

		spaces := 12
		if len(d.Value1) > 0 {
			spaces -= len(d.Value1)
		}

		output += fmt.Sprintf(
			"%s%s%s%s%s%s\n",
			prefix,
			d.Name,
			strings.Repeat(" ", 6),
			d.Value1,
			strings.Repeat(" ", spaces),
			d.Value2,
		)
	}

	return output
}
