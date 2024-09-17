package database

import (
	"github.com/ygrebnov/dbdiff/entity/table"
	"github.com/ygrebnov/dbdiff/query"
)

const DriverPostgresql = "postgres"

type postgresql struct {
	*baseDatabase
}

func newPostgresqlDatabase(uri string, idx uint8) Database {
	d := &postgresql{
		baseDatabase: &baseDatabase{
			schemaQueries: query.Postgresql,
			parser:        table.NewPostgresqlParser(),
			idx:           idx,
		},
	}
	d.createHandler(DriverPostgresql, uri)

	return d
}

func (p *postgresql) GetType() string {
	return Postgresql
}
