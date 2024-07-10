package comparer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygrebnov/dbdiff/entity/field"
	"github.com/ygrebnov/dbdiff/entity/row"
	"github.com/ygrebnov/dbdiff/entity/table"
)

func TestNewFieldsDifference(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name         string
		f1           *field.Field
		f2           *field.Field
		mixedDBTypes bool
		expected     *Difference
	}{
		{
			"equal, same type",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			false,
			&Difference{Name: "f1", Value1: "text not null", Value2: "text not null"},
		},
		{
			"equal, mixed types",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			true,
			&Difference{Name: "f1", Value1: "text", Value2: "text"},
		},
		{
			"not equal, same type",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			&field.Field{Name: "f1", FieldType: "boolean", Attrs: ""},
			false,
			&Difference{Name: "f1", Value1: "text not null", Value2: "boolean"},
		},
		{
			"not equal, mixed types",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			&field.Field{Name: "f1", FieldType: "boolean", Attrs: ""},
			true,
			&Difference{Name: "f1", Value1: "text", Value2: "boolean"},
		},
		{
			"left empty, same type",
			&field.Field{},
			&field.Field{Name: "f1", FieldType: "boolean", Attrs: ""},
			false,
			&Difference{Name: "f1", Value1: "", Value2: "boolean"},
		},
		{
			"left empty, mixed types",
			&field.Field{},
			&field.Field{Name: "f1", FieldType: "boolean", Attrs: ""},
			true,
			&Difference{Name: "f1", Value1: "", Value2: "boolean"},
		},
		{
			"right empty, same type",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			&field.Field{},
			false,
			&Difference{Name: "f1", Value1: "text not null", Value2: ""},
		},
		{
			"right empty, mixed types",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			&field.Field{},
			true,
			&Difference{Name: "f1", Value1: "text", Value2: ""},
		},
		{
			"left nil, same type",
			nil,
			&field.Field{Name: "f1", FieldType: "boolean", Attrs: ""},
			false,
			&Difference{Name: "f1", Value1: "", Value2: "boolean"},
		},
		{
			"left nil, mixed types",
			nil,
			&field.Field{Name: "f1", FieldType: "boolean", Attrs: ""},
			true,
			&Difference{Name: "f1", Value1: "", Value2: "boolean"},
		},
		{
			"right nil, same type",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			nil,
			false,
			&Difference{Name: "f1", Value1: "text not null", Value2: ""},
		},
		{
			"right nil, mixed types",
			&field.Field{Name: "f1", FieldType: "text", Attrs: "not null"},
			nil,
			true,
			&Difference{Name: "f1", Value1: "text", Value2: ""},
		},
		{
			"both nil",
			nil,
			nil,
			false,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := NewFieldsDifference(test.f1, test.f2, test.mixedDBTypes)

			if test.expected != nil {
				require.NotNil(t, d)
				require.Equal(t, *test.expected, *d)
			} else {
				require.Nil(t, d)
			}
		})
	}
}

func TestNewDataRowDifferences(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name                string
		r1                  row.Row
		r2                  row.Row
		tbl                 table.Table
		verbose             bool
		expectedDifferences Differences
		expectedEqual       bool
		expectedErr         string
	}{
		{
			"equal",
			newMockRow("v"),
			newMockRow("v"),
			mockTable1Column,
			false,
			nil,
			true,
			"",
		},
		{
			"not equal",
			newMockRow(s1, s2),
			newMockRow(s2, s1),
			mockTable2Columns,
			false,
			differencesDifferent,
			false,
			"",
		},
		{
			"r1 nil",
			nil,
			newMockRow(s2, s2),
			mockTable2Columns,
			false,
			differencesLeftEmpty,
			false,
			"",
		},
		{
			"r2 nil",
			newMockRow(s1, s1),
			nil,
			mockTable2Columns,
			false,
			differencesRightEmpty,
			false,
			"",
		},
		{
			"both nil",
			nil,
			nil,
			mockTable2Columns,
			false,
			nil,
			true,
			"",
		},
		// verbose
		{
			"equal, verbose",
			newMockRow("v"),
			newMockRow("v"),
			mockTable1Column,
			true,
			Differences{{Name: "v1", Value1: "v", Value2: "v"}},
			true,
			"",
		},
		{
			"not equal, verbose",
			newMockRow(s1, s2),
			newMockRow(s2, s1),
			mockTable2Columns,
			true,
			differencesDifferent,
			false,
			"",
		},
		{
			"r1 nil, verbose",
			nil,
			newMockRow(s2, s2),
			mockTable2Columns,
			true,
			differencesLeftEmpty,
			false,
			"",
		},
		{
			"r2 nil, verbose",
			newMockRow(s1, s1),
			nil,
			mockTable2Columns,
			true,
			differencesRightEmpty,
			false,
			"",
		},
		{
			"both nil, verbose",
			nil,
			nil,
			mockTable2Columns,
			true,
			nil,
			true,
			"",
		},
		// error
		{
			"error",
			newMockRow(s1, s2),
			newMockRow(s2, s1),
			mockTable3Columns,
			false,
			nil,
			false,
			"table columns names number [3] is not equal to values number",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, e, err := NewDataRowDifferences(test.r1, test.r2, test.tbl, test.verbose)

			if test.expectedErr == "" {
				require.NoError(t, err)
				require.Equal(t, test.expectedDifferences, d)
				require.Equal(t, test.expectedEqual, e)
			} else {
				require.ErrorContains(t, err, test.expectedErr)
			}
		})
	}
}

func TestDifferences_Format(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name           string
		differences    Differences
		verbose        bool
		expectedResult string
	}{
		{
			"both empty",
			Differences{},
			false,
			"",
		},
		{
			"left empty",
			differencesLeftEmpty,
			false,
			mockOutput(differencesLeftEmpty, false),
		},
		{
			"right empty",
			differencesRightEmpty,
			false,
			mockOutput(differencesRightEmpty, false),
		},
		{
			"equal",
			differencesEqual,
			false,
			mockOutput(differencesEqual, false),
		},
		{
			"different",
			differencesDifferent,
			false,
			mockOutput(differencesDifferent, false),
		},
		{
			"both empty, verbose",
			Differences{},
			true,
			"",
		},
		{
			"left empty, verbose",
			differencesLeftEmpty,
			true,
			mockOutput(differencesLeftEmpty, true),
		},
		{
			"right empty, verbose",
			differencesRightEmpty,
			true,
			mockOutput(differencesRightEmpty, true),
		},
		{
			"equal, verbose",
			differencesEqual,
			true,
			mockOutput(differencesEqual, true),
		},
		{
			"different, verbose",
			differencesDifferent,
			true,
			mockOutput(differencesDifferent, true),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedResult, test.differences.Format(test.verbose))
		})
	}
}
