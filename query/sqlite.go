package query

import (
	"fmt"
	"strings"
)

type sqlite struct{}

func NewSqliteQuery() Schema {
	return &sqlite{}
}

func (*sqlite) All() string {
	return "SELECT name, sql FROM sqlite_master WHERE type = 'table';"
}

func (*sqlite) One(name string) string {
	return fmt.Sprintf("SELECT sql FROM sqlite_master WHERE name = '%s';", name)
}

func (*sqlite) Excluded(names []string) string {
	return fmt.Sprintf(
		"SELECT name from sqlite_master WHERE type = 'table' AND name not in (%s);",
		strings.Join(names, ","),
	)
}

var Sqlite = NewSqliteQuery()
