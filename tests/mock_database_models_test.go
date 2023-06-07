package tests

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"

	"github.com/ygrebnov/testutils/docker"

	"github.com/ygrebnov/dbdiff/models"
)

const (
	dbTemplate = `BEGIN TRANSACTION;
{{range .Tables}}CREATE TABLE {{.Name}} ({{$nf := len (slice (printf "%*s" (len .Fields) "") 1)}}{{range $i, $f := .Fields}}{{if (ne $i 0)}} {{end}}{{.Name}} {{.FieldType}} {{.Attrs}}{{if (lt $i $nf)}},{{end}}{{end}});{{end}}
{{range .Data}}INSERT INTO {{.Name}} VALUES ({{join .Values "), ("}});{{end}}
COMMIT;`
	dbTypePostgresql = "postgres"
	dbTypeSqlite     = "sqlite"
)

// mockTableData holds mocked table data.
type mockTableData struct {
	Name   string
	Values []string
}

type mockDatabase struct {
	Tables    []models.Table
	Data      []mockTableData
	DBType    string
	Handler   *sql.DB
	Container *docker.DatabaseContainer
	Port      int
	FilePath  string
}

func newMockDatabase(dbType string, container *docker.DatabaseContainer, port *int, filepath string) mockDatabase {
	database := mockDatabase{DBType: dbType}
	switch dbType {
	case dbTypePostgresql:
		database.Container = container
		database.Port = *port
	case dbTypeSqlite:
		database.FilePath = filepath
	}
	return database
}

func (md *mockDatabase) asSQL() string {
	var (
		buff  bytes.Buffer
		funcs = template.FuncMap{"join": strings.Join}
	)
	t := template.Must(template.New("").Funcs(funcs).Parse(dbTemplate))
	if err := t.Execute(&buff, md); err != nil {
		log.Fatal("error generating database sql")
	}
	return buff.String()
}

func (md *mockDatabase) initialize() error {
	var sqlOpenString string
	switch md.DBType {
	case dbTypeSqlite:
		sqlOpenString = md.FilePath
	case dbTypePostgresql:
		sqlOpenString = fmt.Sprintf(
			"postgres://%s:%s@localhost:%d/postgres?sslmode=disable",
			postgresUser,
			postgresPassword,
			md.Port,
		)
	}
	md.Handler, _ = sql.Open(md.DBType, sqlOpenString)
	if _, err := md.Handler.Exec(md.asSQL()); err != nil {
		return err
	}
	return nil
}

func (md *mockDatabase) mockInputString() string {
	s := md.DBType + ":"
	switch md.DBType {
	case dbTypeSqlite:
		s += md.FilePath
	case dbTypePostgresql:
		s += fmt.Sprintf(
			"postgres://%s:%s@localhost:%d/postgres?sslmode=disable",
			postgresUser,
			postgresPassword,
			md.Port,
		)
	}
	return s
}

type difference[T any] struct {
	inD1 T
	inD2 T
}

type exists = difference[bool]
type schemaDifferences = difference[string]
type dataDifferences = difference[string]

// Represents a type holding a table differences in two databases.
type tableDifferences struct {
	rowsNum int // maximum number of rows in a table
	exists
	defaultFields []*models.Field           // default fields in a table
	fields        map[int]exists            // field index in defaultFields -> whether a field exists in table
	fieldTypes    map[int]schemaDifferences // field index in defaultFields -> field types differences
	fieldAttrs    map[int]schemaDifferences // field index in defaultFields -> field attributes differences
	data          map[int]dataDifferences   // map keys length < rowsNum
}

// Represents a type holding two databases schemas and data differences.
type databaseDifferences struct {
	tablesNum         int                // maximum number of tables in a database
	tablesDifferences []tableDifferences // slice size is equal to tablesNum value
}

func newDataRow(i int) string {
	return fmt.Sprintf("'id%d', 'mock_text_value%d', 'TRUE', '2022-12-01 21:00:01'", i, i)
}

func generateDatabases(diffs databaseDifferences, db1 *mockDatabase, db2 *mockDatabase) {
	for t := 0; t < diffs.tablesNum; t++ {
		tableName := "table" + strconv.Itoa(t)
		tableDiff := diffs.tablesDifferences[t]
		table1 := models.Table{Name: tableName}
		table2 := models.Table{Name: tableName}

		// add fields to tables
		for f := 0; f < len(tableDiff.defaultFields); f++ {
			fieldTemplate := tableDiff.defaultFields[f]
			field1 := models.Field{Name: fieldTemplate.Name}
			field2 := models.Field{Name: fieldTemplate.Name}
			if fieldTypeDiff, exists := tableDiff.fieldTypes[f]; exists {
				field1.FieldType = fieldTypeDiff.inD1
				field2.FieldType = fieldTypeDiff.inD2
			}
			if fieldAttrsDiff, exists := tableDiff.fieldAttrs[f]; exists {
				field1.Attrs = fieldAttrsDiff.inD1
				field2.Attrs = fieldAttrsDiff.inD2
			}
			if tableDiff.fields[f].inD1 {
				table1.Fields = append(table1.Fields, &field1)
			}
			if tableDiff.fields[f].inD2 {
				table2.Fields = append(table2.Fields, &field2)
			}
		}

		// add tables to databases
		if tableDiff.inD1 {
			db1.Tables = append(db1.Tables, table1)
		}
		if tableDiff.inD2 {
			db2.Tables = append(db2.Tables, table2)
		}

		// add tables data
		for r := 0; r < tableDiff.rowsNum; r++ {
			tableData1 := mockTableData{Name: tableName}
			tableData2 := mockTableData{Name: tableName}
			tableData1.Values = make([]string, 0, tableDiff.rowsNum)
			tableData2.Values = make([]string, 0, tableDiff.rowsNum)

			if dataDiff, exists := tableDiff.data[r]; exists {
				if len(dataDiff.inD1) > 0 { // to skip missing rows
					tableData1.Values = append(tableData1.Values, dataDiff.inD1)
				}
				if len(dataDiff.inD2) > 0 {
					tableData2.Values = append(tableData2.Values, dataDiff.inD2)
				}
			} else {
				tableData1.Values = append(tableData1.Values, newDataRow(r))
				tableData2.Values = append(tableData2.Values, newDataRow(r))
			}

			if tableDiff.inD1 && len(tableData1.Values) > 0 {
				db1.Data = append(db1.Data, tableData1)
			}
			if tableDiff.inD2 && len(tableData2.Values) > 0 {
				db2.Data = append(db2.Data, tableData2)
			}
		}
	}
}
