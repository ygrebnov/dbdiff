package sql

// Rows is [database/sql.Rows] wrapper. Introduced to facilitate testing.
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
}
