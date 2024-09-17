package database

import (
	"context"
	databaseSQL "database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	_ "github.com/lib/pq" // package is not used directly
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // package is not used directly

	"github.com/ygrebnov/dbdiff/entity/field"
	"github.com/ygrebnov/dbdiff/entity/row"
	"github.com/ygrebnov/dbdiff/entity/table"
	"github.com/ygrebnov/dbdiff/query"
	"github.com/ygrebnov/dbdiff/sql"
)

func TestNewDatabase(t *testing.T) {
	osStat = func(_ string) (any, error) { // mock os.Stat function
		return nil, nil
	}

	var tests = []struct {
		input          string
		expectedDBType string
	}{
		{
			"sqlite:pr.sql",
			Sqlite,
		},
		{
			"postgres:postgres://username:userpassword@hostname:port/dbname",
			Postgresql,
		},
		// TODO: add "unknown database type" scenario.
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			d := New(test.input, 1)
			require.Equal(t, test.expectedDBType, d.GetType())
		})
	}
}

type mockTable struct {
	table.Table
	name string
}

func (m *mockTable) GetName() string {
	return m.name
}

func (m *mockTable) GetFields() []*field.Field {
	return []*field.Field{
		{Name: "field1", FieldType: "text"},
		{Name: "field2", FieldType: "text"},
	}
}

func (m *mockTable) GetFieldsString() string {
	return ""
}

func (m *mockTable) GetPrimaryKey() *field.Field {
	return &field.Field{}
}

func (m *mockTable) SortFields() {}

func newMockTable(name string) table.Table {
	return &mockTable{name: name}
}

// mockData provides mock database data rows via methods defined in [sql.Rows] interface.
type mockData[T interface{}] struct {
	mu        sync.Mutex
	rows      []T
	idx       int
	errIdx    int // index to return a Scan error for.
	errString string
}

func newMockDataWithErr[T interface{}](errIdx int, errString string, rows ...T) *mockData[T] {
	return &mockData[T]{rows: rows, errIdx: errIdx, errString: errString}
}

func newMockData[T interface{}](rows ...T) *mockData[T] {
	return newMockDataWithErr(-1, "", rows...)
}

func (q *mockData[T]) Next() bool {
	q.mu.Lock()

	if q.idx >= len(q.rows) {
		return false
	}

	q.idx++
	return true
}

func (q *mockData[T]) Scan(dest ...interface{}) error {
	defer q.mu.Unlock()

	if q.idx-1 == q.errIdx {
		return fmt.Errorf("%s: %d", q.errString, q.errIdx)
	}

	switch typed := dest[0].(type) {
	case *T:
		*typed = q.rows[q.idx-1]
	}

	return nil
}

type mockRow struct {
	row.Row
	pk string
}

func (m *mockRow) Parse() error {
	return nil
}

func (m *mockRow) GetPKValue() string {
	return m.pk
}

func newMockRow(pk string) row.Row {
	return &mockRow{pk: pk}
}

type mockParser struct {
	err bool
}

func (p *mockParser) Parse(_ table.Table) error {
	if p.err {
		return errors.New("errParsingSchema")
	}
	return nil
}

var (
	table1 = newMockTable("table1")
	table2 = newMockTable("table2")
	row1   = newMockRow("1")
	row2   = newMockRow("2")
)

func TestBaseDatabase_GetTables(t *testing.T) {
	var tests = []struct {
		name             string
		tables           []table.Table
		errGetData       bool
		errNewTable      bool
		errParsingSchema bool
	}{
		{
			name:   "no errors",
			tables: []table.Table{table1, table2},
		},
		{
			name:       "getData error",
			tables:     []table.Table{table1, table2},
			errGetData: true,
		},
		{
			name:        "newTable error",
			tables:      []table.Table{table1, table2},
			errNewTable: true,
		},
		{
			name:             "schema parsing error",
			tables:           []table.Table{table1, table2},
			errParsingSchema: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			getData = func(_ *databaseSQL.DB, _ string, _ ...interface{}) (sql.Rows, error) {
				if test.errGetData {
					return nil, errors.New("errGetData")
				}

				return newMockData(test.tables...), nil
			}

			newUnparsedTable = func(rows sql.Rows) (table.Table, error) {
				if test.errNewTable {
					return nil, errors.New("errNewTable")
				}

				var tbl table.Table

				err := rows.Scan(&tbl)
				require.NoError(t, err)

				return tbl, nil
			}

			d := &baseDatabase{
				schemaQueries: query.NewPostgresqlQuery(),
				parser:        &mockParser{err: test.errParsingSchema},
			}

			tables, errs := d.GetTables(context.Background())

			expectedNames := make(map[string]struct{}, len(test.tables))
			for i := range test.tables {
				expectedNames[test.tables[i].GetName()] = struct{}{}
			}

			actualNames := make(map[string]struct{}, len(expectedNames))

			if test.errGetData {
				// expect GetTables to immediately return an error via errors channel.
				require.Nil(t, tables)
				require.ErrorContains(t, <-errs, "errGetData")
			} else {
				// listen to both channels, tables and errors, but until tables is open.
				for tbl := range tables {
					select {
					case err := <-errs:
						switch {
						case err != nil && test.errNewTable:
							require.ErrorContains(t, err, "errNewTable")
						case err != nil && !test.errParsingSchema:
							require.ErrorContains(t, err, "errParsingSchema")
						default:
							t.Fail()
						}
					default:
					}

					actualNames[tbl.GetName()] = struct{}{}
				}

				if test.errNewTable || test.errParsingSchema {
					require.Less(t, len(actualNames), len(test.tables))
				} else {
					require.Equal(t, len(test.tables), len(actualNames))
				}
			}
		})
	}
}

func TestBaseDatabase_GetTable(t *testing.T) {
	var tests = []struct {
		name        string
		tblName     string
		tbl         table.Table
		errNoRows   bool
		errGetData  bool
		errNewTable bool
	}{
		{
			name:    "no errors",
			tblName: "table1",
			tbl:     table1,
		},
		{
			name:      "noRows error",
			tblName:   "table1",
			tbl:       table1,
			errNoRows: true,
		},
		{
			name:       "getData error",
			tblName:    "table1",
			tbl:        table1,
			errGetData: true,
		},
		{
			name:        "newTable error",
			tblName:     "table1",
			tbl:         table1,
			errNewTable: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			getData = func(_ *databaseSQL.DB, _ string, _ ...interface{}) (sql.Rows, error) {
				switch {
				case test.errNoRows:
					return nil, databaseSQL.ErrNoRows
				case test.errGetData:
					return nil, errors.New("errGetData")
				default:
					return newMockData(test.tbl), nil
				}
			}

			newTable = func(rows sql.Rows, _ table.Parser) (table.Table, error) {
				if test.errNewTable {
					return nil, errors.New("errNewTable")
				}

				var tbl table.Table

				err := rows.Scan(&tbl)
				require.NoError(t, err)

				return tbl, nil
			}

			d := &baseDatabase{schemaQueries: query.NewPostgresqlQuery()}

			tbl, err := d.GetTable(test.tblName)

			switch {
			case test.errNoRows:
				require.Nil(t, tbl)
				require.ErrorContains(t, err, fmt.Sprintf("Table '%s' does not exist in database", test.tblName))

			case test.errGetData:
				require.Nil(t, tbl)
				require.ErrorContains(t, err, fmt.Sprintf("cannot get table '%s' schema from database", test.tblName))

			case test.errNewTable:
				require.Nil(t, tbl)
				require.ErrorContains(t, err, "errNewTable")

			default:
				require.Equal(t, test.tblName, tbl.GetName())
			}
		})
	}
}

func TestBaseDatabase_GetTableNames(t *testing.T) {
	var tests = []struct {
		name          string
		tableNames    []string
		errGetData    bool
		errGetNameIdx int
	}{
		{
			name:          "no errors",
			tableNames:    []string{"table1", "table2"},
			errGetNameIdx: -1,
		},
		{
			name:          "getData error",
			tableNames:    []string{"table1", "table2"},
			errGetData:    true,
			errGetNameIdx: -1,
		},
		{
			name:          "get name error",
			tableNames:    []string{"table1", "table2"},
			errGetNameIdx: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			getData = func(_ *databaseSQL.DB, _ string, _ ...interface{}) (sql.Rows, error) {
				if test.errGetData {
					return nil, errors.New("errGetData")
				}

				return newMockDataWithErr(test.errGetNameIdx, "errGetName", test.tableNames...), nil
			}

			d := &baseDatabase{schemaQueries: query.NewPostgresqlQuery()}

			names, errs, err := d.GetTableNames([]string{})

			switch {
			case test.errGetData:
				require.ErrorContains(t, err, "errGetData")
				require.Nil(t, names)
				require.Nil(t, errs)

			case test.errGetNameIdx != -1:
				require.Nil(t, err)

				require.NotNil(t, errs)
				require.Contains(t, errs[0], "errGetName")
				require.Contains(t, errs[0], strconv.Itoa(test.errGetNameIdx))

				require.NotNil(t, names)
				require.Less(t, len(names), len(test.tableNames))

			default:
				require.Nil(t, err)
				require.Empty(t, errs)
				require.ElementsMatch(t, names, test.tableNames)
			}
		})
	}
}

func TestBaseDatabase_GetTableRows(t *testing.T) {
	var tests = []struct {
		name         string
		rows         []row.Row
		errGetData   bool
		errNewRowIdx int
	}{
		{
			name:         "no errors",
			rows:         []row.Row{row1, row2},
			errNewRowIdx: -1,
		},
		{
			name:         "empty data",
			rows:         []row.Row{},
			errNewRowIdx: -1,
		},
		{
			name:         "getData error",
			rows:         []row.Row{row1, row2},
			errGetData:   true,
			errNewRowIdx: -1,
		},
		{
			name:         "newRow error",
			rows:         []row.Row{row1, row2},
			errNewRowIdx: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			getData = func(_ *databaseSQL.DB, _ string, _ ...interface{}) (sql.Rows, error) {
				if test.errGetData {
					return nil, errors.New("errGetData")
				}

				return newMockDataWithErr(test.errNewRowIdx, "errNewRow", test.rows...), nil
			}

			newRow = func(_ []*field.Field, rows sql.Rows, _ bool) (row.Row, error) {
				var r row.Row

				err := rows.Scan(&r)

				return r, err
			}

			d := &baseDatabase{schemaQueries: query.NewPostgresqlQuery()}

			rows, errs, err := d.getTableRows(table1, "")

			actual := make(map[row.Row]struct{}, len(test.rows))

			if test.errGetData {
				// expect GetTableRows to immediately return an error.
				require.Nil(t, rows)
				require.ErrorContains(t, err, "errGetData")
			} else {
				for range len(test.rows) {
					select {
					case r := <-rows:
						require.NotEqual(t, strconv.Itoa(test.errNewRowIdx+1), r.GetPKValue(), "should have received an error")

						actual[r] = struct{}{}

					case e := <-errs:
						require.NotNil(t, e)
						require.ErrorContains(t, e, "errNewRow")
						require.ErrorContains(t, e, strconv.Itoa(test.errNewRowIdx))
					}
				}

				if test.errNewRowIdx == -1 {
					require.Equal(t, len(test.rows), len(actual))
				} else {
					require.Less(t, len(actual), len(test.rows))
				}
			}
		})
	}
}

func TestBaseDatabase_GetTableRow(t *testing.T) {
	var tests = []struct {
		name       string
		pk         string
		r          row.Row
		errNoRows  bool
		errGetData bool
		errNewRow  bool
	}{
		{
			name: "no errors",
			pk:   "1",
			r:    row1,
		},
		{
			name:      "noRows error",
			pk:        "1",
			r:         row1,
			errNoRows: true,
		},
		{
			name:       "getData error",
			pk:         "1",
			r:          row1,
			errGetData: true,
		},
		{
			name:      "newRow error",
			pk:        "1",
			r:         row1,
			errNewRow: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			getData = func(_ *databaseSQL.DB, _ string, _ ...interface{}) (sql.Rows, error) {
				switch {
				case test.errNoRows:
					return nil, databaseSQL.ErrNoRows
				case test.errGetData:
					return nil, errors.New("errGetData")
				default:
					return newMockData(test.r), nil
				}
			}

			newRow = func(_ []*field.Field, rows sql.Rows, _ bool) (row.Row, error) {
				if test.errNewRow {
					return nil, errors.New("errNewRow")
				}

				var r row.Row

				err := rows.Scan(&r)
				require.NoError(t, err)

				return r, nil
			}

			d := &baseDatabase{schemaQueries: query.NewPostgresqlQuery()}

			r, err := d.GetTableRow(table1, test.pk)

			switch {
			case test.errNoRows:
				require.Nil(t, r)
				require.Nil(t, err)

			case test.errGetData:
				require.Nil(t, r)
				require.ErrorContains(t, err, fmt.Sprintf("cannot fetch '%s' table data row", table1.GetName()))

			case test.errNewRow:
				require.Nil(t, r)
				require.ErrorContains(t, err, "errNewRow")

			default:
				require.Equal(t, test.pk, r.GetPKValue())
			}
		})
	}
}
