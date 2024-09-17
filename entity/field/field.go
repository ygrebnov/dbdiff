package field

import (
	"fmt"
	"strings"
)

// Field holds field object attributes.
type Field struct {
	Name         string
	FieldType    string
	IsPrimaryKey bool
	// Attrs holds field object attributes other than name and type.
	// Attributes differ in databases of different types.
	Attrs string
}

// New returns a new Field from the given raw string.
// An error is returned in case name or type attributes values cannot be resolved.
func New(raw string) (*Field, error) {
	attrs := strings.Split(raw, " ")

	if len(attrs) < 2 {
		return nil, fmt.Errorf("invalid field: %s", raw)
	}

	f := &Field{
		Name:         attrs[0],
		FieldType:    attrs[1],
		IsPrimaryKey: strings.Contains(raw, "primary key"),
	}

	if len(attrs) > 2 {
		f.Attrs = strings.Join(attrs[2:], " ")
	}

	return f, nil
}

// IsEqual returns true if this and given field types and attributes are equal.
// In case the databases are of different types, fields attributes comparison is omitted.
func (f *Field) IsEqual(other *Field, mixedDBTypes bool) bool {
	if other == nil {
		return false
	}

	return !(f.FieldType != other.FieldType || !mixedDBTypes && f.Attrs != other.Attrs)
}
