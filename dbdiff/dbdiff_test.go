package dbdiff

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockDifferences []Difference

func (md mockDifferences) asShortOutput() string {
	output := "\n    Field   Database1   Database2\n"
	for _, diff := range md {
		if len(diff.Value1) == 0 {
			diff.Value1 = "         "
		}
		output += fmt.Sprintf("    %s      %s   %s\n", diff.Name, diff.Value1, diff.Value2)
	}
	return output
}

func (md mockDifferences) asLongOutput() string {
	output := "\n    Field   Database1   Database2\n"
	for _, diff := range md {
		if diff.Value1 == diff.Value2 {
			output += "  ="
		} else {
			output += "  x"
		}
		if len(diff.Value1) == 0 {
			diff.Value1 = "         "
		}
		output += fmt.Sprintf(" %s      %s   %s\n", diff.Name, diff.Value1, diff.Value2)
	}
	return output
}

var (
	mockV0Ctx       = context.Background()
	mockV1Ctx       = context.WithValue(context.Background(), VerboseContextKey, true)
	mockV2Ctx       = context.WithValue(context.Background(), VVerboseContextKey, true)
	mockV3Ctx       = context.WithValue(context.Background(), VVVerboseContextKey, true)
	mockS1          = "mocked_s1"
	mockS2          = "mocked_s2"
	mockValues      = []string{mockS1, mockS2}
	mockEmptyValues = []string{"", ""}
	mockS1Empty     = mockDifferences{{"v1", "", mockS2}, {"v2", "", mockS2}}
	mockS2Empty     = mockDifferences{{"v1", mockS1, mockS2}, {"v2", mockS2, mockS1}}
	mockDiff        = mockDifferences{{"v1", mockS1, ""}, {"v2", mockS1, ""}}
	mockColumns     = []string{"pk", "v1", "v2"}
)

func TestGetDifferences(t *testing.T) {
	var tests = []struct {
		name                string
		ctx                 context.Context
		fieldsNum           int
		values1             []string
		values2             []string
		columns             []string
		expectedDifferences []Difference
		expectedEqual       bool
	}{
		{"v0_equal", mockV0Ctx, 2, mockValues, mockValues, mockColumns, []Difference{}, true},
		{"v1_equal", mockV1Ctx, 2, mockValues, mockValues, mockColumns, []Difference{}, true},
		{"v2_equal", mockV2Ctx, 2, mockValues, mockValues, mockColumns, []Difference{}, true},
		{
			"v3_equal", mockV3Ctx, 2, mockValues, mockValues, mockColumns,
			[]Difference{
				{Name: "v1", Value1: mockS1, Value2: mockS1},
				{Name: "v2", Value1: mockS2, Value2: mockS2},
			},
			true,
		},
		{
			"v0_not_equal", mockV0Ctx, 2,
			mockValues,
			[]string{mockS1, mockS1},
			mockColumns,
			[]Difference{{Name: "v2", Value1: mockS2, Value2: mockS1}},
			false,
		},
		{
			"v1_not_equal", mockV1Ctx, 2,
			mockValues,
			[]string{mockS1, mockS1},
			mockColumns,
			[]Difference{{Name: "v2", Value1: mockS2, Value2: mockS1}},
			false,
		},
		{
			"v2_not_equal", mockV2Ctx, 2,
			mockValues,
			[]string{mockS1, mockS1},
			mockColumns,
			[]Difference{{Name: "v2", Value1: mockS2, Value2: mockS1}},
			false,
		},
		{
			"v3_not_equal", mockV3Ctx, 2,
			mockValues,
			[]string{mockS1, mockS1},
			mockColumns,
			[]Difference{
				{Name: "v1", Value1: mockS1, Value2: mockS1},
				{Name: "v2", Value1: mockS2, Value2: mockS1},
			},
			false,
		},
		{"v0_empty", mockV0Ctx, 2, mockEmptyValues, mockEmptyValues, mockColumns, []Difference{}, true},
		{"v1_empty", mockV1Ctx, 2, mockEmptyValues, mockEmptyValues, mockColumns, []Difference{}, true},
		{"v2_empty", mockV2Ctx, 2, mockEmptyValues, mockEmptyValues, mockColumns, []Difference{}, true},
		{
			"v3_empty", mockV3Ctx, 2, mockEmptyValues, mockEmptyValues, mockColumns,
			[]Difference{{"v1", "", ""}, {"v2", "", ""}},
			true,
		},
	}
	for _, tt := range tests {
		var (
			actualDifferences = []Difference{}
			actualEqual       = true
		)
		t.Run(tt.name, func(t *testing.T) {
			getDifferences(
				tt.ctx, tt.fieldsNum, tt.values1, tt.values2, tt.columns, &actualDifferences, &actualEqual,
			)
			require.Equal(t, tt.expectedDifferences, actualDifferences)
			require.Equal(t, tt.expectedEqual, actualEqual)
		})
	}
}

func TestFormatDifferences(t *testing.T) {
	var tests = []struct {
		name           string
		ctx            context.Context
		differences    []Difference
		expectedResult string
	}{
		{"v0_no_differences", mockV0Ctx, []Difference{}, ""},
		{"v0_left_empty", mockV0Ctx, mockS1Empty, mockS1Empty.asShortOutput()},
		{"v0_right_empty", mockV0Ctx, mockS2Empty, mockS2Empty.asShortOutput()},
		{"v0_different", mockV0Ctx, mockDiff, mockDiff.asShortOutput()},
		{"v1_no_differences", mockV1Ctx, []Difference{}, ""},
		{"v1_left_empty", mockV1Ctx, mockS1Empty, mockS1Empty.asShortOutput()},
		{"v1_right_empty", mockV1Ctx, mockS2Empty, mockS2Empty.asShortOutput()},
		{"v1_different", mockV1Ctx, mockDiff, mockDiff.asShortOutput()},
		{"v2_no_differences", mockV2Ctx, []Difference{}, ""},
		{"v2_left_empty", mockV2Ctx, mockS1Empty, mockS1Empty.asLongOutput()},
		{"v2_right_empty", mockV2Ctx, mockS2Empty, mockS2Empty.asLongOutput()},
		{"v2_different", mockV2Ctx, mockDiff, mockDiff.asLongOutput()},
		{"v3_no_differences", mockV3Ctx, []Difference{}, ""},
		{"v3_left_empty", mockV3Ctx, mockS1Empty, mockS1Empty.asLongOutput()},
		{"v3_right_empty", mockV3Ctx, mockS2Empty, mockS2Empty.asLongOutput()},
		{"v3_different", mockV3Ctx, mockDiff, mockDiff.asLongOutput()},
	}
	for _, tt := range tests {
		var actualResult string
		t.Run(tt.name, func(t *testing.T) {
			formatDifferences(tt.ctx, &tt.differences, &actualResult)
			require.Equal(t, tt.expectedResult, actualResult)
		})
	}
}
