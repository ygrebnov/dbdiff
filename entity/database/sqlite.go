package database

import (
	"log"

	"github.com/ygrebnov/dbdiff/entity/table"
	"github.com/ygrebnov/dbdiff/query"
)

const DriverSqlite = "sqlite"

type sqlite struct {
	*baseDatabase
}

func newSqliteDatabase(uri string, idx uint8) Database {
	if _, err := osStat(uri); err != nil {
		log.Fatalf("database file %s does not exist\n", uri)
	}

	d := &sqlite{
		baseDatabase: &baseDatabase{
			schemaQueries: query.Sqlite,
			parser:        table.NewSqliteParser(),
			idx:           idx,
		},
	}
	d.createHandler(DriverSqlite, uri)

	return d
}

func (p *sqlite) GetType() string {
	return Sqlite
}
