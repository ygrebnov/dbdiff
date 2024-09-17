package comparer

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/ygrebnov/dbdiff/entity/field"
	"github.com/ygrebnov/dbdiff/entity/row"
	"github.com/ygrebnov/dbdiff/entity/table"
)

const (
	differencesTemplate = "    Field\tDatabase1\tDatabase2\n{{range .}}    " +
		"{{.Name}}\t{{.Value1}}\t{{.Value2}}\n{{end}}"
	verboseDifferencesTemplate = "    Field\tDatabase1\tDatabase2\n{{range .}}  " +
		"{{ if eq .Value1 .Value2 }}={{ else }}x{{ end }} {{.Name}}\t{{.Value1}}\t{{.Value2}}\n{{end}}"
)

// Difference is a type capable of holding an identified by name entity values in compared entities.
// For example, a name of a table field with values in database1 and database2.
type Difference struct {
	Name, Value1, Value2 string
}

type Differences []*Difference

// Format returns differences formatted depending on the verbosity level.
func (d Differences) Format(verbose bool) string {
	if len(d) == 0 {
		return ""
	}

	var (
		buff bytes.Buffer
		out  strings.Builder
	)

	t := differencesTemplate
	if verbose {
		t = verboseDifferencesTemplate
	}

	tmpl := template.Must(template.New("").Parse(t))
	w := tabwriter.NewWriter(&buff, 5, 0, 3, ' ', 0)
	if err := tmpl.Execute(w, d); err != nil {
		out.WriteString(fmt.Sprintf("\n    error formatting differences: %v", d))
		return out.String()
	}
	w.Flush()

	out.Write(buff.Bytes())

	return out.String()
}

// NewFieldsDifference returns two field.Field objects schemas Difference.
func NewFieldsDifference(f1, f2 *field.Field, mixedDBTypes bool) *Difference {
	var name string

	switch {
	case f1 == nil && f2 == nil:
		return nil
	case f1 == nil:
		f1 = &field.Field{}
		name = f2.Name
	case f2 == nil:
		f2 = &field.Field{}
		name = f1.Name
	case len(f1.Name) > 0:
		name = f1.Name
	default:
		name = f2.Name
	}

	v1 := f1.FieldType
	v2 := f2.FieldType
	if !mixedDBTypes {
		space1 := ""
		space2 := ""

		if len(f1.Attrs) > 0 {
			space1 = " "
		}
		if len(f2.Attrs) > 0 {
			space2 = " "
		}

		v1 = fmt.Sprintf("%s%s%s", f1.FieldType, space1, f1.Attrs)
		v2 = fmt.Sprintf("%s%s%s", f2.FieldType, space2, f2.Attrs)
	}

	return &Difference{
		Name:   name,
		Value1: v1,
		Value2: v2,
	}
}

func NewDataRowDifferences(r1, r2 row.Row, t table.Table, verbose bool) (Differences, bool, error) {
	if r1 == nil && r2 == nil {
		return nil, true, nil
	}

	names := t.GetFieldNames()

	var values1, values2 []string

	if r1 != nil {
		values1 = r1.GetValues()
	}

	if r2 != nil {
		values2 = r2.GetValues()
	}

	if (r1 != nil && len(names) != len(values1)) || (r2 != nil && len(names) != len(values2)) {
		return nil, false, fmt.Errorf(
			"%s table columns names number [%d] is not equal to values number v1=[%d], v2=[%d]",
			t.GetName(),
			len(names),
			len(values1),
			len(values2),
		)
	}

	equalRow := true
	var d Differences

	if r1 == nil || r2 == nil {
		equalRow = false
		d = make(Differences, 0, len(names))
	}

	for i := range names {
		var v1, v2 string

		if r1 != nil {
			v1 = values1[i]
		}

		if r2 != nil {
			v2 = values2[i]
		}

		equal := v1 == v2

		if equalRow && !equal {
			equalRow = false
		}

		if !equal || verbose {
			d = append(d, &Difference{Name: names[i], Value1: v1, Value2: v2})
		}
	}

	return d, equalRow, nil
}
