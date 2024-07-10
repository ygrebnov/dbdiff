package table

type Parser interface {
	Parse(table Table) error
}
