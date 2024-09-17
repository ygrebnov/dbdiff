package comparer

import (
	"context"
	"sync"

	_ "github.com/lib/pq"  // package is not used directly
	_ "modernc.org/sqlite" // package is not used directly

	"github.com/ygrebnov/dbdiff/entity/database"
	"github.com/ygrebnov/dbdiff/entity/table"
)

// databaseComparer is a type capable of comparing two databases by analyzing their schemas and data.
type databaseComparer struct {
	verbosity int // TODO: replace by config.

	d1, d2       database.Database
	mixedDBTypes bool // TODO: replace by config.

	results *results
}

// NewDatabaseComparer returns a new database comparer given two resources as connection strings or data file paths.
func NewDatabaseComparer(resource1, resource2 string, verbosity int) Comparer {
	d1 := database.New(resource1, 1)
	d2 := database.New(resource2, 2)

	return &databaseComparer{
		verbosity: verbosity,

		d1:           d1,
		d2:           d2,
		mixedDBTypes: d1.GetType() != d2.GetType(),

		results: newResults(verbosity),
	}
}

// Compare performs two databases schemas and data comparison.
func (c *databaseComparer) Compare(ctx context.Context) {
	wg := &sync.WaitGroup{}

	tables := make(chan table.Table)
	done := make(chan bool, 1)

	go c.compareSchemas(ctx, tables, done)
	go func() {
		<-done
		close(tables)
	}()

	for t := range tables {
		wg.Add(1)
		go func(t table.Table) {
			defer wg.Done()
			c.compareData(t)
		}(t)
	}

	wg.Wait()
}
