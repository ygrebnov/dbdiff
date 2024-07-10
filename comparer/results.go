package comparer

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/ygrebnov/dbdiff/entity/table"
)

type results struct {
	schemas map[uuid.UUID]chan string // key: table id.
	data    map[uuid.UUID]chan string // key: table id.

	verbosity int
}

func newResults(verbosity int) *results {
	return &results{
		schemas: make(map[uuid.UUID]chan string),
		data:    make(map[uuid.UUID]chan string),

		verbosity: verbosity,
	}
}

func (r *results) addSchema(id uuid.UUID) chan string {
	r.schemas[id] = make(chan string, 1024)

	return r.schemas[id]
}

func (r *results) addData(id uuid.UUID) chan string {
	r.data[id] = make(chan string, 1024)

	return r.data[id]
}

func (r *results) getData(id uuid.UUID) chan<- string {
	return r.data[id]
}

// output prints out table comparison results.
// Table comparison results channel is drained.
func (r *results) output(t table.Table, equal bool) {
	id := t.GetID()

	schemas, bSchemas := r.schemas[id]
	if bSchemas {
		bSchemas = len(schemas) > 0 // a channel may exist but empty.

		for range len(schemas) {
			fmt.Println(<-schemas)
		}

		close(schemas)

		delete(r.schemas, id)
	}

	data, bData := r.data[id]

	// data differences.
	switch {
	case equal && r.verbosity < 3 && r.verbosity > 0:
		fmt.Println("  data differences: none")
	case equal && r.verbosity == 3:
		fmt.Println("  data differences:")

	case !equal && !bSchemas && bData:
		fmt.Printf("Table %s data differences:\n", t.GetName())
	case !equal && bSchemas && bData:
		fmt.Println("  data differences:")
	}

	if bData {
		for range len(data) {
			fmt.Println(<-data)
		}

		close(data)

		delete(r.data, id)
	}
}
