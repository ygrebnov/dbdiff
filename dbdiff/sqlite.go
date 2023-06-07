package dbdiff

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/ygrebnov/dbdiff/models"
)

// sqliteDatabase defines methods applicable to an SQLite database. Implements [models.DatabaseType] interface.
type sqliteDatabase struct{}

// newSqliteDatabase returns a new sqliteDatabase object.
func newSqliteDatabase() models.DatabaseType {
	return &sqliteDatabase{}
}

func (*sqliteDatabase) Name() string {
	return "sqlite"
}

func (*sqliteDatabase) Driver() string {
	return "sqlite"
}

func (*sqliteDatabase) Parse(table *models.Table) {
	sql := regexp.MustCompile(`(?s)\(.+\)`).FindString(table.Schema)
	fields := regexp.MustCompile(`(?s)(\S+[^,]+?(,|$))`).FindAllString(sql[1:len(sql)-1], -1)
	if len(fields) == 0 {
		panic("cannot parse table schema") //TODO: replace with returning error
	}

	for _, field := range fields {
		field = strings.TrimFunc(
			strings.ToLower(field),
			func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) },
		)
		if !strings.HasPrefix(field, "create table") &&
			!strings.HasPrefix(field, "foreign key") &&
			!strings.HasPrefix(field, ")") &&
			len(field) > 0 {
			(*table).AddField(field)
		}
	}
}

func (*sqliteDatabase) QueryAll() string {
	return "SELECT name, sql FROM sqlite_master WHERE type = 'table';"
}

func (*sqliteDatabase) QueryOne(name string) string {
	return fmt.Sprintf("SELECT sql FROM sqlite_master WHERE name = '%s';", name)
}

func (*sqliteDatabase) QueryExcluded(names []string) string {
	return fmt.Sprintf(
		"SELECT name from sqlite_master WHERE type = 'table' AND name not in (%s);",
		strings.Join(names, ","),
	)
}

var sqlite = newSqliteDatabase()
