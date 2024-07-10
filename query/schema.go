package query

// Schema provides database tables names and schemas queries.
type Schema interface {
	// All returns a query selecting all tables names and schemas.
	All() string
	// One returns a query selecting given table schema.
	One(name string) string
	// Excluded returns a query selecting names of all tables except excluded.
	Excluded(names []string) string
}
