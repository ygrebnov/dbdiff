package query

import (
	"fmt"
	"strings"

	"github.com/ygrebnov/dbdiff/entity/table"
)

type data struct{}

// All returns a query for fetching the given table rows.
// Each row contains table primary key field value followed by all other fields values sorted in alphabetical order.
func (d *data) All(t table.Table) string {
	return fmt.Sprintf("SELECT %s, %s FROM %s;", t.GetPrimaryKey().Name, t.GetFieldsString(), t.GetName())
}

// One returns a query for fetching a row for a given table primary key value.
func (d *data) One(t table.Table, pkValue string) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = '%s';",
		t.GetFieldsString(),
		t.GetName(),
		t.GetPrimaryKey().Name,
		pkValue,
	)
}

// Excluded returns a query string for fetching table data excluding rows with the given primary key values.
func (d *data) Excluded(t table.Table, keys []string) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s NOT IN (%s);",
		t.GetFieldsString(), t.GetName(), t.GetPrimaryKey().Name, strings.Join(keys, ","),
	)
}

var Data = &data{}
