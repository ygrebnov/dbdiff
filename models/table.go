package models

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// Table holds table object attributes.
type Table struct {
	Name       string
	Schema     string
	PrimaryKey *Field
	Fields     []*Field
	// FieldNameIndex is a map of field name to its index in fields slice.
	FieldNameIndex map[string]int
	DB             *Database
	// ComparisonResult is a cumulative result of table comparison in two databases.
	ComparisonResult string
}

// ID returns table identifier used in comparison output.
func (t *Table) ID() string {
	return fmt.Sprintf("Table %s", t.Name)
}

// GetFieldsFromRow gets table fields by parsing table schema provided in a row.
func (t *Table) GetFieldsFromRow(row *sql.Rows) error {
	if err := row.Scan(&(t.Name), &(t.Schema)); err != nil {
		return err
	}
	t.DB.DBType.Parse(t)
	return nil
}

// GetFields gets given table fields.
func (t *Table) GetFields(name string) error {
	optionalSQL := new(sql.NullString)
	if err := t.DB.Handler.QueryRow(t.DB.DBType.QueryOne(name)).Scan(optionalSQL); err != nil {
		return err
	}
	if optionalSQL.Valid {
		t.Schema = optionalSQL.String
	} else {
		// postgres driver does not return this error if main query returns empty result
		return sql.ErrNoRows
	}
	t.DB.DBType.Parse(t)
	return nil
}

// AddField adds a field to table by parsing given raw string.
func (t *Table) AddField(rawField string) {
	attrs := strings.Split(rawField, " ")
	f := Field{
		Name:       attrs[0],
		FieldType:  attrs[1],
		PrimaryKey: strings.Contains(rawField, "primary key"),
	}
	if len(attrs) > 2 {
		f.Attrs = strings.Join(attrs[2:], " ")
	}
	if f.PrimaryKey {
		t.PrimaryKey = &f
	}
	t.Fields = append(t.Fields, &f)
	i := len(t.Fields) - 1
	t.FieldNameIndex[attrs[0]] = i
}

// SortFields sorts table fields in alphabetical order.
func (t *Table) SortFields() {
	if !sort.SliceIsSorted(
		t.Fields,
		func(i, j int) bool { return t.Fields[i].Name < t.Fields[j].Name },
	) {
		sort.SliceStable(
			t.Fields,
			func(i, j int) bool { return t.Fields[i].Name < t.Fields[j].Name },
		)
	}
}

// FieldsSQL returns concatenated table fields names sorted in alphabetical order.
func (t *Table) FieldsSQL() string {
	var s string
	t.SortFields()
	for i, f := range t.Fields {
		if i == 0 {
			s += f.Name
		} else {
			s += ", " + f.Name
		}
	}
	return s
}

// ParseRawSQLValues converts retrieved from database raw values of type NullString (optional strings) into string ones.
// Slice of input raw values corresponds to the table one row. Length of input raw values must be equal to the length of the output string values
// and the number of table fields. Order of fields in the input raw values slice must be the same as in the output string values one and must correspond to
// the table alphabetically ordered field names.
// Values of fields of type boolean are transformed to lower case.
func (t *Table) ParseRawSQLValues(rawValues *[]any, values *[]string) {
	for i, el := range *rawValues {
		if ns, ok := el.(*sql.NullString); ok && ns.Valid {
			(*values)[i] = ns.String
			if t.Fields[i].FieldType == "boolean" {
				(*values)[i] = strings.ToLower((*values)[i])
			}
		}
	}
}

// QueryDataAll returns table data query string. The query retrieves table primary key field name followed by all the fields sorted in alphabetical order.
func (t *Table) QueryDataAll() string {
	return fmt.Sprintf("SELECT %s, %s FROM %s;", t.PrimaryKey.Name, t.FieldsSQL(), t.Name)
}

// QueryDataExcluded returns a query string for fetching table data excluding rows with the given primary key values.
func (t *Table) QueryDataExcluded(epk *[]string) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s NOT IN (%s);",
		t.FieldsSQL(), t.Name, t.PrimaryKey.Name, strings.Join(*epk, ","),
	)
}
