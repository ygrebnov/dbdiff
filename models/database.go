package models

import (
	"database/sql"
	"fmt"
)

// DatabaseType defines a database type by its name, driver, the tables schemas and data queries,
// as well as the way of how the tables schemas are parsed.
type DatabaseType interface {
	// Name returns database type name.
	Name() string
	// Driver returns database type driver name.
	Driver() string
	// Parse parses table schema.
	Parse(table *Table)
	// QueryAll returns a query selecting all tables names and schemas.
	QueryAll() string
	// QueryOne returns a query selecting given table schema.
	QueryOne(name string) string
	// QueryExcluded returns a query selecting names of all tables except excluded.
	QueryExcluded(names []string) string
}

// Database holds database object attributes.
type Database struct {
	DBType  DatabaseType
	Handler *sql.DB
	// URI is a connection string or a data file path.
	URI string
}

// FetchDataRowFromTable fetches a row for a given table primary key value.
func (d *Database) FetchDataRowFromTable(table *Table, name string, data []any) error {
	return d.Handler.QueryRow(
		fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = '%s';",
			table.FieldsSQL(),
			table.Name,
			table.PrimaryKey.Name,
			name,
		),
	).Scan(data...)
}
