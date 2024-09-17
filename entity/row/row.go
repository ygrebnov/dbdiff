package row

import (
	databaseSQL "database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ygrebnov/dbdiff/entity/field"
	"github.com/ygrebnov/dbdiff/sql"
)

type Row interface {
	Parse() error
	GetPKValue() string
	GetValues() []string
}

type row struct {
	raw     []interface{} // unparsed values.
	fields  []*field.Field
	pkValue string
	values  []string
}

// New returns a new Row with unparsed raw fields.
func New(fields []*field.Field, data sql.Rows, withPk bool) (Row, error) {
	n := len(fields)
	r := &row{fields: fields, values: make([]string, n)}

	if withPk {
		n++
	}

	r.raw = make([]interface{}, n)
	for i := range n {
		r.raw[i] = new(databaseSQL.NullString)
	}

	if err := data.Scan(r.raw...); err != nil {
		return nil, fmt.Errorf("error fetching data from database: %w", err)
	}

	return r, nil
}

func (r *row) Parse() error {
	hasPk := len(r.values) != len(r.raw)

	for i, v := range r.raw {
		ns := v.(*databaseSQL.NullString)

		j := i
		if hasPk {
			j--
		}

		switch {
		case !ns.Valid:
			return errors.New("invalid row data")

		case hasPk && i == 0:
			r.pkValue = ns.String

		case r.fields[j].FieldType == "boolean":
			r.values[j] = strings.ToLower(ns.String)

		default:
			r.values[j] = ns.String
		}
	}

	return nil
}

func (r *row) GetPKValue() string {
	return r.pkValue
}

func (r *row) GetValues() []string {
	return r.values
}

func NewError(fields []*field.Field) Row {
	values := make([]string, len(fields))
	for i := range fields {
		values[i] = "error retrieving data"
	}

	return &row{values: values}
}
