package dbdiff

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	_ "github.com/lib/pq"  // package is not used directly
	_ "modernc.org/sqlite" // package is not used directly

	"github.com/ygrebnov/dbdiff/models"
)

// databaseComparer is a type capable of comparing two databases by analyzing their schemas and data.
// Implements [Comparer] interface.
type databaseComparer struct{}

// NewDatabaseComparer returns a new databaseComparer object.
func NewDatabaseComparer() Comparer {
	return &databaseComparer{}
}

// parse gets database type and URI from input string.
func (dc *databaseComparer) parse(input string, database any) {
	var (
		isSqlite, isPostgresql bool
		uri                    string
	)
	d, _ := database.(*models.Database)
	if _, uri, isSqlite = strings.Cut(input, fmt.Sprintf("%s:", sqlite.Name())); isSqlite {
		d.DBType = sqlite
		d.URI = uri
		if _, err := osStat(uri); err != nil {
			log.Fatalf("database file %s does not exist\n", uri)
		}
	} else if _, uri, isPostgresql = strings.Cut(input, fmt.Sprintf("%s:", postgresql.Name())); isPostgresql {
		d.DBType = postgresql
		d.URI = uri
	} else {
		log.Fatalln("unknown database type")
	}
}

// createHandler creates a handler for given database.
func (dc *databaseComparer) createHandler(database *models.Database) {
	var err error
	database.Handler, err = sql.Open(database.DBType.Driver(), database.URI)
	if err != nil {
		log.Fatalln("cannot create database handler:", err)
	}
}

// Compare performs two databases schemas and data comparison.
func (dc *databaseComparer) Compare(ctx context.Context, input1, input2 string) {
	var d1, d2 models.Database

	// get database types and URIs from database strings
	dc.parse(input1, &d1)
	dc.parse(input2, &d2)

	// create database handlers
	dc.createHandler(&d1)
	dc.createHandler(&d2)

	ctx = context.WithValue(ctx, mixedDBTypesContextKey, d1.DBType.Name() != d2.DBType.Name())

	// create channels
	tablesChannel := make(chan *models.Table)
	resultsChannel := make(chan string)
	schemasDoneChannel := make(chan bool, 1)
	dataDoneChannel := make(chan bool, 1)

	wg := &sync.WaitGroup{}

	go compareDatabaseSchemas(ctx, d1, d2, tablesChannel, resultsChannel, schemasDoneChannel)
	go watchSchemaComparisonCompletion(schemasDoneChannel, tablesChannel)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range tablesChannel {
			wg.Add(1)
			go func(t *models.Table) {
				defer wg.Done()
				compareTablesData(ctx, d1, d2, t, resultsChannel)
			}(t)
		}
	}()

	go watchDataComparisonCompletion(wg, dataDoneChannel, resultsChannel)

	for result := range resultsChannel {
		fmt.Println(result)
	}

	<-dataDoneChannel
}

// watchSchemaComparisonCompletion closes tables channel on schemas comparison completion.
func watchSchemaComparisonCompletion(done chan bool, tables chan *models.Table) {
	<-done
	close(tables)
}

// watchDataComparisonCompletion closes results channel on data comparison completion.
func watchDataComparisonCompletion(wg *sync.WaitGroup, dataComparisonDone chan bool, results chan string) {
	wg.Wait()
	close(results)
	dataComparisonDone <- true
}

// compareDatabaseSchemas compares two given databases schemas.
// nolint: funlen
func compareDatabaseSchemas(
	ctx context.Context,
	d1, d2 models.Database,
	tables chan *models.Table,
	results chan string,
	done chan bool,
) {
	var (
		comparedTables []string // keeps track of already compared database tables
		rows1, rows2   *sql.Rows
		err            error
	)
	mixedDBTypes := fromContext(ctx, mixedDBTypesContextKey, true)
	verbose := fromContext(ctx, VVerboseContextKey, false)
	verbose = fromContext(ctx, VVVerboseContextKey, verbose) || verbose

	rows1, err = d1.Handler.Query(d1.DBType.QueryAll())
	if err != nil {
		log.Panicln("cannot get database1 schema")
	}
	defer rows1.Close()
	for rows1.Next() {
		var (
			t1, t2      models.Table
			differences []Difference // holds differences in table schemas in databases
		)
		equal := true
		t1.DB = &d1
		t2.DB = &d2
		t1.FieldNameIndex = make(map[string]int)
		t2.FieldNameIndex = make(map[string]int)

		if err = t1.GetFieldsFromRow(rows1); err != nil {
			results <- fmt.Sprintf("%s: error retrieving schema from database1: %q", t1.ID(), err)
			continue
		}

		comparedTables = append(comparedTables, fmt.Sprintf("'%s'", t1.Name))

		// get table schema from the second database
		if err = t2.GetFields(t1.Name); err != nil {
			if err == sql.ErrNoRows {
				results <- t1.ID() + ": does not exist in database2"
			} else {
				results <- t1.ID() + fmt.Sprintf(": error retrieving schema from database2: %q", err)
			}
			continue
		}

		// compare schemas
		visited := make(map[int]struct{}) // indices of t2 fields which exist in t1
		t1.ComparisonResult = fmt.Sprintf("%s:\n  schema differences:", t1.ID())
		for _, f := range t1.Fields {
			var t2Type, t2Attr string
			j, exists := t2.FieldNameIndex[f.Name] // check if a field with the given name exists in t2
			if exists {
				visited[j] = struct{}{}
				t2Type = t2.Fields[j].FieldType
				t2Attr = t2.Fields[j].Attrs
			}
			if !exists || f.FieldType != t2Type || (!mixedDBTypes && f.Attrs != t2Attr) {
				equal = false
			}
			if !exists || f.FieldType != t2Type || (!mixedDBTypes && f.Attrs != t2Attr) || verbose {
				if mixedDBTypes {
					differences = append(differences, Difference{
						Name:   f.Name,
						Value1: f.FieldType,
						Value2: t2Type,
					})
				} else {
					differences = append(differences, Difference{
						Name:   f.Name,
						Value1: fmt.Sprintf("%s %s", f.FieldType, f.Attrs),
						Value2: fmt.Sprintf("%s %s", t2Type, t2Attr),
					})
				}
			}
		}

		// process the rest of t2 fields left uncompared
		for i, f := range t2.Fields {
			if _, skip := visited[i]; !skip { // check indices of t2 fields which do not exist in t1
				if mixedDBTypes {
					differences = append(differences, Difference{
						Name:   f.Name,
						Value2: f.FieldType,
					})
				} else {
					differences = append(differences, Difference{
						Name:   f.Name,
						Value2: fmt.Sprintf("%s %s", f.FieldType, f.Attrs),
					})
				}
				equal = false
			}
		}

		formatDifferences(ctx, &differences, &(t1.ComparisonResult))
		if !equal {
			results <- t1.ComparisonResult
			continue
		}

		// schemas are equal, continue with data comparison
		if !verbose {
			t1.ComparisonResult += " none"
		}
		tables <- &t1
	}

	// process tables from the second database which are not in comparedTables slice
	rows2, err = d2.Handler.Query(d2.DBType.QueryExcluded(comparedTables))
	if err != nil {
		log.Panicln("cannot get database2 schema")
	}
	defer rows2.Close()
	for rows2.Next() {
		var t models.Table
		if err = rows2.Scan(&t.Name); err != nil {
			log.Panicln("cannot get database2 schema")
		}
		results <- t.ID() + ": does not exist in database1"
	}

	done <- true
}

// compareTablesData compares given table data in two databases.
// nolint: funlen, gocyclo
func compareTablesData(ctx context.Context, d1, d2 models.Database, t *models.Table, results chan string) {
	var (
		rows1, rows2 *sql.Rows
		columns      []string // contains pk at 0 position
		comparedPks  []string // keeps track of already compared primary key values
		err          error
	)

	fieldsNum := len(t.Fields)
	equal := true                                               // comparison result
	result := fmt.Sprintf("Table %s data differences:", t.Name) // comparison details

	verbose := fromContext(ctx, VerboseContextKey, false)
	vverbose := fromContext(ctx, VVerboseContextKey, false)
	vvverbose := fromContext(ctx, VVVerboseContextKey, false)

	if verbose || vverbose || vvverbose {
		result = fmt.Sprintf("%s\n  data differences:", t.ComparisonResult)
	}

	if rows1, err = d1.Handler.Query(t.QueryDataAll()); err != nil && err != sql.ErrNoRows {
		results <- result + " cannot get data from database1"
		return
	}
	defer rows1.Close()

	columns, err = rows1.Columns()
	if err != nil {
		results <- result + " cannot get data from database1"
		return
	}

	line := 0

	for rows1.Next() {
		line++

		var differences []Difference // holds differences in table row data in databases
		equalLine := true
		rawValues1 := make([]any, fieldsNum+1) // contains primary key value at 0 position
		rawValues2 := make([]any, fieldsNum)
		values1 := make([]string, fieldsNum)
		values2 := make([]string, fieldsNum)

		for i := 0; i < fieldsNum+1; i++ {
			rawValues1[i] = new(sql.NullString)
			if i < fieldsNum {
				rawValues2[i] = new(sql.NullString)
			}
		}
		var pkValue string // primary key value

		if err := rows1.Scan(rawValues1...); err != nil {
			result += fmt.Sprintf("\n    line %d: error fetching data from database1", line)
			equal = false
			continue
		}
		for i, el := range rawValues1 {
			if ns, ok := el.(*sql.NullString); ok && ns.Valid {
				if i == 0 {
					pkValue = ns.String
					comparedPks = append(comparedPks, fmt.Sprintf("'%s'", pkValue))
				} else {
					values1[i-1] = ns.String
					if t.Fields[i-1].FieldType == "boolean" {
						values1[i-1] = strings.ToLower(values1[i-1])
					}
				}
			}
		}

		// fetch data for a given pk value from database2
		if err = d2.FetchDataRowFromTable(t, pkValue, rawValues2); err != nil {
			result += fmt.Sprintf("\n  line %d (%s=%s):", line, t.PrimaryKey.Name, pkValue)
			if err != sql.ErrNoRows {
				for i := 0; i < fieldsNum; i++ {
					values2[i] = "error retrieving data"
				}
			}
			getDifferences(ctx, fieldsNum, values1, values2, columns, &differences, &equal)
			formatDifferences(ctx, &differences, &result)
			continue
		}

		t.ParseRawSQLValues(&rawValues2, &values2)
		getDifferences(ctx, fieldsNum, values1, values2, columns, &differences, &equalLine)
		if !equalLine || vvverbose {
			result += fmt.Sprintf("\n  line %d (%s=%s):", line, t.PrimaryKey.Name, pkValue)
		}
		formatDifferences(ctx, &differences, &result)

		if !equalLine {
			equal = false
		}
	}

	// Add remaining data from database2
	if rows2, err = d2.Handler.Query(t.QueryDataExcluded(&comparedPks)); err != nil && err != sql.ErrNoRows {
		results <- fmt.Sprintf("%s cannot get data from database2: %q", result, err)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		line++

		var differences []Difference
		rawValues2 := make([]any, fieldsNum)
		values1 := make([]string, fieldsNum)
		values2 := make([]string, fieldsNum)
		for i := 0; i < fieldsNum; i++ {
			rawValues2[i] = new(sql.NullString)
		}

		if err = rows2.Scan(rawValues2...); err != nil {
			result += fmt.Sprintf("\n  line %d: error fetching data from database2", line)
			equal = false
			continue
		}
		t.ParseRawSQLValues(&rawValues2, &values2)
		result += fmt.Sprintf("\n  line %d:", line)
		getDifferences(ctx, fieldsNum, values1, values2, columns, &differences, &equal)
		formatDifferences(ctx, &differences, &result)
	}

	if equal && (verbose || vverbose) {
		result += " none"
	}

	if !equal || verbose || vverbose || vvverbose {
		results <- result
	}
}

// osStat is used to simplify testing
var osStat = func(name string) (any, error) {
	return os.Stat(name)
}
