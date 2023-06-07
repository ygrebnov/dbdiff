package dbdiff

import (
	"bytes"
	"context"
	"fmt"
	"text/tabwriter"
	"text/template"
)

// Comparer is a type capable of comparing two entities.
type Comparer interface {
	Compare(ctx context.Context, entity1 string, entity2 string)
	parse(input string, entity any)
}

// contextKey is a context key type.
type contextKey string

// Difference is a type capable of holding an identified by name entity values in compared entities.
// For example, a name of a table field with values in database1 and database2.
type Difference struct {
	Name, Value1, Value2 string
}

const (
	differencesTemplate = "    Field\tDatabase1\tDatabase2\n{{range .}}    " +
		"{{.Name}}\t{{.Value1}}\t{{.Value2}}\n{{end}}"
	verboseDifferencesTemplate = "    Field\tDatabase1\tDatabase2\n{{range .}}  " +
		"{{ if eq .Value1 .Value2 }}={{ else }}x{{ end }} {{.Name}}\t{{.Value1}}\t{{.Value2}}\n{{end}}"
	VerboseContextKey      = contextKey("verbose")
	VVerboseContextKey     = contextKey("vverbose")
	VVVerboseContextKey    = contextKey("vvverbose")
	mixedDBTypesContextKey = contextKey("mixedDBTypes")
)

// getDifferences compares values from two databases and puts their differences in a slice.
func getDifferences(
	ctx context.Context,
	fieldsNum int,
	values1, values2, columns []string,
	differences *[]Difference,
	overallEqual *bool,
) {
	// equal values are added to the differences result only at the third level of verbosity.
	verbose := fromContext(ctx, VVVerboseContextKey, false)

	for i := 0; i < fieldsNum; i++ {
		equal := values1[i] == values2[i]
		// by default, the overall equality is true. It changes only if any individual values are different.
		if !equal {
			*overallEqual = false
		}
		if !equal || verbose {
			*differences = append(*differences, Difference{Name: columns[i+1], Value1: values1[i], Value2: values2[i]})
		}
	}
}

// formatDifferences formats comparison differences as a table and adds them to the resulting output string.
// Formatting template depends on the verbosity level.
func formatDifferences(ctx context.Context, differences *[]Difference, result *string) {
	if len(*differences) > 0 {
		var buff bytes.Buffer

		t := differencesTemplate
		if fromContext(ctx, VVerboseContextKey, false) || fromContext(ctx, VVVerboseContextKey, false) {
			t = verboseDifferencesTemplate
		}

		tmpl := template.Must(template.New("").Parse(t))
		w := tabwriter.NewWriter(&buff, 5, 0, 3, ' ', 0)
		if err := tmpl.Execute(w, differences); err != nil {
			*result += fmt.Sprintf("\n    error formatting differences: %v", differences)
		}
		w.Flush()
		*result += fmt.Sprintf("\n%s", buff.String())
	}
}

// fromContext retrieves value for a given key from context and returns it.
func fromContext[T any](ctx context.Context, key contextKey, defaultValue T) T {
	value := defaultValue
	if v, ok := ctx.Value(key).(T); ok {
		value = v
	}
	return value
}
