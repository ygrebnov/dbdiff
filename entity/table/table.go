package table

import (
	"cmp"
	databaseSQL "database/sql"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/ygrebnov/dbdiff/entity/field"
	"github.com/ygrebnov/dbdiff/sql"
)

// Table represents a database table.
type Table interface {
	AddField(f *field.Field)
	SortFields()
	GetID() uuid.UUID
	GetName() string
	GetSchema() string
	GetPrimaryKey() *field.Field
	GetFields() []*field.Field
	GetFieldNames() []string
	// GetFieldByName returns the Field and its index, given its name.
	GetFieldByName(name string) (int, *field.Field)
	GetFieldByIndex(index int) *field.Field
	GetFieldsString() string
}

type table struct {
	mu sync.Mutex

	id           uuid.UUID
	name         string
	schema       string
	primaryKey   *field.Field
	fields       []*field.Field
	fieldNames   []string
	fieldsString string
	// fieldsIndexes is a map of field name to its index in fields slice.
	fieldsIndexes map[string]int
}

// newUnparsed returns a new Table.
// primary means 'a primary table for comparison'. In this sense, secondary table does not need neither id, nor name.
func newUnparsed(row sql.Rows, primary bool) (Table, error) {
	t := &table{fieldsIndexes: make(map[string]int)}
	var (
		err          error
		name, schema databaseSQL.NullString
	)

	switch primary {
	case true:
		t.id = uuid.New()
		err = row.Scan(&name, &schema)
	default:
		err = row.Scan(&schema)
	}

	if err != nil {
		return nil, err
	}

	if !schema.Valid || (primary && !name.Valid) {
		return nil, nil
	}

	t.schema = schema.String
	if primary {
		t.name = name.String
	}

	return t, nil
}

func New(row sql.Rows, parser Parser) (Table, error) {
	t, e := newUnparsed(row, false)
	switch {
	case e != nil:
		return nil, e
	case t == nil:
		return nil, nil
	}

	if e = parser.Parse(t); e != nil {
		return nil, fmt.Errorf("error parsing '%s' table schema: %q", t.GetName(), e)
	}

	t.SortFields()

	return t, nil
}

func NewUnparsed(row sql.Rows) (Table, error) {
	return newUnparsed(row, true)
}

func (t *table) SortFields() {
	slices.SortStableFunc(t.fields, func(a, b *field.Field) int { return cmp.Compare(a.Name, b.Name) })

	t.fieldNames = make([]string, len(t.fields))
	for i, f := range t.fields {
		t.fieldNames[i] = f.Name
		t.fieldsIndexes[f.Name] = i
	}

	t.fieldsString = strings.Join(t.fieldNames, ",")
}

// AddField adds a field to table.
func (t *table) AddField(f *field.Field) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if f.IsPrimaryKey {
		t.primaryKey = f
	}
	t.fields = append(t.fields, f)
}

// GetID returns table UUID.
func (t *table) GetID() uuid.UUID {
	return t.id
}

// GetName returns table name.
func (t *table) GetName() string {
	return t.name
}

// GetSchema returns table schema.
func (t *table) GetSchema() string {
	return t.schema
}

func (t *table) GetPrimaryKey() *field.Field {
	return t.primaryKey
}

func (t *table) GetFields() []*field.Field {
	return t.fields
}

func (t *table) GetFieldNames() []string {
	return t.fieldNames
}

func (t *table) GetFieldByName(name string) (int, *field.Field) {
	if idx, ok := t.fieldsIndexes[name]; ok {
		return idx, t.fields[idx]
	}
	return -1, nil
}

func (t *table) GetFieldByIndex(index int) *field.Field {
	if index < 0 || index >= len(t.fields) {
		return nil
	}

	return t.fields[index]
}

func (t *table) GetFieldsString() string {
	return t.fieldsString
}
