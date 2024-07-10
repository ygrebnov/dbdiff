package query

// Results is sql.Rows wrapper. Introduced to facilitate testing.
type Results interface {
	Next() bool
	Scan(dest ...interface{}) error
}
