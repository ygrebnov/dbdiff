package database

import (
	"context"
	databaseSQL "database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/ygrebnov/dbdiff/entity/row"
	"github.com/ygrebnov/dbdiff/entity/table"
	"github.com/ygrebnov/dbdiff/query"
	"github.com/ygrebnov/dbdiff/sql"
)

const (
	Postgresql = "postgres"
	Sqlite     = "sqlite"
)

type Database interface {
	GetType() string
	GetTables(ctx context.Context) (<-chan table.Table, <-chan error)
	GetTableNames(excluded []string) ([]string, []string, error)
	GetTable(name string) (table.Table, error)
	GetTableRowsAll(t table.Table) (<-chan row.Row, <-chan error, error)
	GetTableRowsExcluding(t table.Table, keys []string) (<-chan row.Row, <-chan error, error)
	GetTableRow(t table.Table, key string) (row.Row, error)
}

func New(resource string, idx uint8) Database {
	if _, uri, ok := strings.Cut(resource, fmt.Sprintf("%s:", Sqlite)); ok {
		return newSqliteDatabase(uri, idx)
	}

	if _, uri, ok := strings.Cut(resource, fmt.Sprintf("%s:", Postgresql)); ok {
		return newPostgresqlDatabase(uri, idx)
	}

	log.Fatalln("unknown database type")
	return nil
}

type baseDatabase struct {
	idx           uint8
	handler       *databaseSQL.DB
	schemaQueries query.Schema
	parser        table.Parser
}

// createHandler creates a database handler given a driver and a URI.
func (d *baseDatabase) createHandler(driver, uri string) {
	var err error
	d.handler, err = databaseSQL.Open(driver, uri)
	if err != nil {
		log.Fatalln("cannot create database handler:", err)
	}
}

// testing hooks.
var (
	osStat = func(name string) (any, error) {
		return os.Stat(name)
	}
	newTable         = table.New
	newUnparsedTable = table.NewUnparsed
	newRow           = row.New
	getData          = func(handler *databaseSQL.DB, query string, args ...interface{}) (sql.Rows, error) {
		return handler.Query(query, args...)
	}
)

func (d *baseDatabase) GetTables(ctx context.Context) (<-chan table.Table, <-chan error) {
	errs := make(chan error, 1)

	data, err := getData(d.handler, d.schemaQueries.All())

	if err != nil {
		errs <- fmt.Errorf("cannot get database%d schema: %w", d.idx, err)

		return nil, errs
	}

	tables := make(chan table.Table)

	g, ctx := errgroup.WithContext(ctx)

	// database/sql.Rows Next() and Scan() methods must be called in pairs.
	// Cannot create a queue of objects to be scanned only, because a subsequent Next() call closes the previous Rows.
	for data.Next() {
		t, e := newUnparsedTable(data)
		if e != nil {
			errs <- e
			break
		}

		g.Go(func() error {
			if e = d.parser.Parse(t); e != nil {
				return fmt.Errorf("error parsing '%s' table schema: %q", t.GetName(), err)
			}

			t.SortFields()

			select {
			case tables <- t:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})
	}

	go func() {
		if err = g.Wait(); err != nil {
			errs <- err
		}

		close(tables)
	}()

	return tables, errs
}

func (d *baseDatabase) GetTable(name string) (table.Table, error) {
	schema, err := getData(d.handler, d.schemaQueries.One(name))

	switch {
	case err != nil && errors.Is(err, databaseSQL.ErrNoRows):
		return nil, fmt.Errorf("Table '%s' does not exist in database%d", name, d.idx)

	case err != nil:
		return nil, fmt.Errorf("cannot get table '%s' schema from database%d: %w", name, d.idx, err)

	case schema == nil || !schema.Next():
		return nil, fmt.Errorf("Table '%s' does not exist in database%d", name, d.idx)
	}

	t, err := newTable(schema, d.parser)
	switch {
	case err != nil:
		return nil, fmt.Errorf("cannot get table '%s' schema from database%d: %w", name, d.idx, err)
	case t == nil:
		return nil, fmt.Errorf("Table '%s' does not exist in database%d", name, d.idx)
	default:
		return t, nil
	}
}

func (d *baseDatabase) GetTableNames(excluded []string) (names []string, errs []string, err error) {
	data, err := getData(d.handler, d.schemaQueries.Excluded(excluded))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get database%d schema: %w", d.idx, err)
	}

	for data.Next() {
		var name string
		if e := data.Scan(&name); e != nil {
			errs = append(errs, e.Error())
		} else {
			names = append(names, name)
		}
	}

	return
}

func (d *baseDatabase) GetTableRowsAll(t table.Table) (<-chan row.Row, <-chan error, error) {
	return d.getTableRows(t, query.Data.All(t))
}

func (d *baseDatabase) GetTableRowsExcluding(t table.Table, keys []string) (<-chan row.Row, <-chan error, error) {
	return d.getTableRows(t, query.Data.Excluded(t, keys))
}

// getTableRows queries given table.Table data from the database and returns it as row.Row objects.
// Method returns two channels, one for row.Row objects and one for errors or an error.
// Created row.Row objects and/or errors are sent to the corresponding channels.
// The channels are closed after all queried data have been processed.
func (d *baseDatabase) getTableRows(t table.Table, q string) (<-chan row.Row, <-chan error, error) {
	data, err := getData(d.handler, q)
	if err != nil && !errors.Is(err, databaseSQL.ErrNoRows) {
		return nil, nil, fmt.Errorf("cannot get '%s' table data from database%d: %w", t.GetName(), d.idx, err)
	}

	rows := make(chan row.Row)
	errs := make(chan error)

	var wg sync.WaitGroup

	for data.Next() {
		r, e := newRow(t.GetFields(), data, true)
		wg.Add(1)

		go func(r row.Row, e error) {
			defer wg.Done()

			if e != nil {
				errs <- e
				return
			}

			if e = r.Parse(); e != nil {
				errs <- e
				return
			}

			rows <- r
		}(r, e)
	}

	go func() {
		wg.Wait()

		close(errs)
		close(rows)
	}()

	return rows, errs, nil
}

// GetTableRow returns given table.Table data row.Row for a given primary key value.
func (d *baseDatabase) GetTableRow(t table.Table, key string) (row.Row, error) {
	data, err := getData(d.handler, query.Data.One(t, key))

	switch {
	case err != nil && errors.Is(err, databaseSQL.ErrNoRows):
		// nil will be considered as empty value in comparer.
		return nil, nil

	case err != nil:
		return nil, fmt.Errorf(
			"cannot fetch '%s' table data row for %s='%s' from database%d: %w",
			t.GetName(),
			t.GetPrimaryKey().Name,
			key,
			d.idx,
			err,
		)

	case data == (&databaseSQL.Rows{}) || !data.Next():
		return nil, nil
	}

	r, err := newRow(t.GetFields(), data, false)
	if err != nil {
		return nil, err
	}

	if err = r.Parse(); err != nil {
		return nil, err
	}

	return r, nil
}
