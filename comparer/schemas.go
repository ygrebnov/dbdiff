package comparer

import (
	"context"
	"fmt"
	"log"

	"github.com/ygrebnov/dbdiff/entity/table"
)

// compareSchemas compares tables schemas in two databases.
func (c *databaseComparer) compareSchemas(ctx context.Context, tc chan<- table.Table, done chan<- bool) {
	var compared []string // holds names of already compared database tables.

	tables, errs := c.d1.GetTables(ctx)
	if tables == nil {
		log.Fatalln(<-errs) // TODO: continue with d2. Stop on DB query execution err.
	}

	for t1 := range tables {
		select {
		case err := <-errs:
			fmt.Println(err.Error())
			continue
		default:
		}

		name := t1.GetName()
		compared = append(compared, fmt.Sprintf("'%s'", name))

		// get table schema from the second database.
		t2, err := c.d2.GetTable(name)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		equal := c.compareFields(t1, t2)

		if !equal {
			c.results.output(t1, false)
			continue
		}

		// schemas are equal, continue with data comparison.
		tc <- t1
	}

	// process tables from the second database which are not in compared slice.
	names, errors, err := c.d2.GetTableNames(compared)
	if err != nil {
		log.Fatalln(err)
	}

	for _, name := range names {
		fmt.Printf("Table '%s' does not exist in database1\n", name)
	}

	for _, e := range errors {
		fmt.Println(e) // TODO: make optional, put into config.
	}

	done <- true
}

func (c *databaseComparer) compareFields(t1, t2 table.Table) bool {
	out := c.results.addSchema(t1.GetID())
	d := Differences{}
	visited := make(map[int]struct{}) // indexes of t2 table fields which exist in t1.
	equal := true

	for _, f1 := range t1.GetFields() {
		j, f2 := t2.GetFieldByName(f1.Name) // check if a field with the given name exists in t2.
		if j > -1 {
			visited[j] = struct{}{}
		}

		switch {
		case j == -1 || !f1.IsEqual(f2, c.mixedDBTypes):
			equal = false
			d = append(d, NewFieldsDifference(f1, f2, c.mixedDBTypes))

		case c.verbosity > 1:
			d = append(d, NewFieldsDifference(f1, f2, c.mixedDBTypes))
		}
	}

	// process the rest of t2 fields left not compared.
	for i, f2 := range t2.GetFields() {
		if _, ok := visited[i]; !ok { // check indexes of other table fields which do not exist in this one.
			d = append(d, NewFieldsDifference(nil, f2, c.mixedDBTypes))
			equal = false
		}
	}

	switch {
	case (!equal || c.verbosity > 1) && len(d) > 0:
		out <- fmt.Sprintf("Table %s:\n  schema differences:", t1.GetName())
		out <- d.Format(c.verbosity > 1)

	case c.verbosity == 1:
		out <- fmt.Sprintf("Table %s:\n  schema differences: none", t1.GetName())
	}

	return equal
}
