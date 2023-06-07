package models

// Field holds field object attributes.
type Field struct {
	Name       string
	FieldType  string
	PrimaryKey bool
	// Attrs holds field object attributes. Attributes differ in databases of different types.
	Attrs string
}
