package tests

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ygrebnov/testutils/docker"

	"github.com/ygrebnov/dbdiff/comparer"
)

type testContainers []docker.DatabaseContainer

var ctx = context.Background()

func captureOutput() func() (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	done := make(chan error, 1)

	save := os.Stdout
	os.Stdout = w

	var buf strings.Builder

	go func() {
		_, err := io.Copy(&buf, r)
		r.Close()
		done <- err
	}()

	return func() (string, error) {
		os.Stdout = save
		w.Close()
		err := <-done
		return buf.String(), err
	}
}

func setUp(t *testing.T, containers testContainers) {
	tearDown(t, containers)
	require.NoError(t, os.Mkdir("test", 0777))
	for _, container := range containers {
		require.NoError(t, container.CreateStart(ctx))
	}
}

func resetEnv(t *testing.T, containers testContainers) {
	for _, container := range containers {
		require.NoError(t, container.ResetDatabase(ctx))
	}
	require.NoError(t, os.RemoveAll("test"))
	require.NoError(t, os.Mkdir("test", 0777))
}

func tearDown(t *testing.T, containers testContainers) {
	for _, container := range containers {
		require.NoError(t, container.StopRemove(ctx))
	}
	require.NoError(t, os.RemoveAll("test"))
}

func TestCompareDatabases(t *testing.T) {
	tests := []struct {
		name           string
		db1            mockDatabase
		db2            mockDatabase
		containers     testContainers
		dbDifferences  databaseDifferences
		verbosity      int
		expectedOutput string
	}{
		// verbosity 0
		{
			"sqlite_sqlite_equal_verbosity0",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsEqual1Table,
			0,
			"",
		},
		{
			"sqlite_sqlite_left_table_absent_verbosity0",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsLeftTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_sqlite_right_table_absent_verbosity0",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsRightTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_sqlite_schema_different_verbosity0",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsSchema1Table,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_text_field        text not null unique   text not null\n" +
				"    mock_timestamp_field   boolean                timestamp not null default current_timestamp\n\n",
		},
		{
			"sqlite_sqlite_differences_verbosity0",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsMixed2Tables,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"    mock_timestamp_field   timestamp not null default current_timestamp   \n" +
				"    mock_boolean_field                                                    boolean\n\n" +
				"Table table1 data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"sqlite_postgres_equal_verbosity0",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsEqual1Table,
			0,
			"",
		},
		{
			"sqlite_postgres_left_table_absent_verbosity0",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_postgres_right_table_absent_verbosity0",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsRightTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_postgres_schema_different_verbosity0",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsSchema1Table,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"sqlite_postgres_differences_verbosity0",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsMixed2Tables,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   timestamp   \n" +
				"    mock_boolean_field                 boolean\n\n" +
				"Table table1 data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_sqlite_equal_verbosity0",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsEqual1Table,
			0,
			"",
		},
		{
			"postgres_sqlite_left_table_absent_verbosity0",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsLeftTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_sqlite_right_table_absent_verbosity0",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsRightTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_sqlite_schema_different_verbosity0",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsSchema1Table,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"postgres_sqlite_differences_verbosity0",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsMixed2Tables,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   timestamp   \n" +
				"    mock_boolean_field                 boolean\n\n" +
				"Table table1 data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_postgres_equal_verbosity0",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsEqual1Table,
			0,
			"",
		},
		{
			"postgres_postgres_left_table_absent_verbosity0",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_postgres_right_table_absent_verbosity0",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsRightTableAbsent1Table,
			0,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_postgres_schema_different_verbosity0",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsSchema1Table,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   boolean     timestamp without time zone\n\n",
		},
		{
			"postgres_postgres_differences_verbosity0",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsMixed2Tables,
			0,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"    mock_timestamp_field   timestamp without time zone   \n" +
				"    mock_boolean_field                                   boolean\n\n" +
				"Table table1 data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		// verbosity 1
		{
			"sqlite_sqlite_equal_verbosity1",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsEqual1Table,
			1,
			"Table table0:\n  schema differences: none\n  data differences: none\n",
		},
		{
			"sqlite_sqlite_left_table_absent_verbosity1",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsLeftTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_sqlite_right_table_absent_verbosity1",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsRightTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_sqlite_schema_different_verbosity1",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsSchema1Table,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_text_field        text not null unique   text not null\n" +
				"    mock_timestamp_field   boolean                timestamp not null default current_timestamp\n\n",
		},
		{
			"sqlite_sqlite_differences_verbosity1",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsMixed2Tables,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"    mock_timestamp_field   timestamp not null default current_timestamp   \n" +
				"    mock_boolean_field                                                    boolean\n\n" +
				"Table table1:\n" +
				"  schema differences: none\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"sqlite_postgres_equal_verbosity1",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsEqual1Table,
			1,
			"Table table0:\n  schema differences: none\n  data differences: none\n",
		},
		{
			"sqlite_postgres_left_table_absent_verbosity1",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_postgres_right_table_absent_verbosity1",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsRightTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_postgres_schema_different_verbosity1",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsSchema1Table,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"sqlite_postgres_differences_verbosity1",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsMixed2Tables,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   timestamp   \n" +
				"    mock_boolean_field                 boolean\n\n" +
				"Table table1:\n" +
				"  schema differences: none\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_sqlite_equal_verbosity1",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsEqual1Table,
			1,
			"Table table0:\n  schema differences: none\n  data differences: none\n",
		},
		{
			"postgres_sqlite_left_table_absent_verbosity1",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsLeftTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_sqlite_right_table_absent_verbosity1",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsRightTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_sqlite_schema_different_verbosity1",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsSchema1Table,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"postgres_sqlite_differences_verbosity1",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsMixed2Tables,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   timestamp   \n" +
				"    mock_boolean_field                 boolean\n\n" +
				"Table table1:\n" +
				"  schema differences: none\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_postgres_equal_verbosity1",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsEqual1Table,
			1,
			"Table table0:\n  schema differences: none\n  data differences: none\n",
		},
		{
			"postgres_postgres_left_table_absent_verbosity1",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_postgres_right_table_absent_verbosity1",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsRightTableAbsent1Table,
			1,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_postgres_schema_different_verbosity1",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsSchema1Table,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"    mock_timestamp_field   boolean     timestamp without time zone\n\n",
		},
		{
			"postgres_postgres_differences_verbosity1",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsMixed2Tables,
			1,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"    mock_timestamp_field   timestamp without time zone   \n" +
				"    mock_boolean_field                                   boolean\n\n" +
				"Table table1:\n" +
				"  schema differences: none\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"    mock_boolean_field   true               false\n" +
				"    mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"    mock_boolean_field     true                   \n" +
				"    mock_id_field          id2                    \n" +
				"    mock_text_field        mock_text_value2       \n" +
				"    mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		// verbosity 2
		{
			"sqlite_sqlite_equal_verbosity2",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsEqual1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"  = mock_boolean_field     boolean                                        boolean\n" +
				"  = mock_id_field          text primary key                               text primary key\n" +
				"  = mock_text_field        text not null unique                           text not null unique\n" +
				"  = mock_timestamp_field   timestamp not null default current_timestamp   timestamp not null default current_timestamp\n\n" +
				"  data differences: none\n",
		},
		{
			"sqlite_sqlite_left_table_absent_verbosity2",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsLeftTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_sqlite_right_table_absent_verbosity2",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsRightTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_sqlite_schema_different_verbosity2",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsSchema1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     boolean                boolean\n" +
				"  = mock_id_field          text primary key       text primary key\n" +
				"  x mock_text_field        text not null unique   text not null\n" +
				"  x mock_timestamp_field   boolean                timestamp not null default current_timestamp\n\n",
		},
		{
			"sqlite_sqlite_differences_verbosity2",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsMixed2Tables,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"  = mock_id_field          text primary key                               text primary key\n" +
				"  = mock_text2_field       text not null                                  text not null\n" +
				"  = mock_text_field        text not null unique                           text not null unique\n" +
				"  x mock_timestamp_field   timestamp not null default current_timestamp   \n" +
				"  x mock_boolean_field                                                    boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"  = mock_boolean_field     boolean                                        boolean\n" +
				"  = mock_id_field          text primary key                               text primary key\n" +
				"  = mock_text_field        text not null unique                           text not null unique\n" +
				"  = mock_timestamp_field   timestamp not null default current_timestamp   timestamp not null default current_timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"  x mock_boolean_field   true               false\n" +
				"  x mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"sqlite_postgres_equal_verbosity2",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsEqual1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences: none\n",
		},
		{
			"sqlite_postgres_left_table_absent_verbosity2",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_postgres_right_table_absent_verbosity2",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsRightTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_postgres_schema_different_verbosity2",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsSchema1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"sqlite_postgres_differences_verbosity2",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsMixed2Tables,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text2_field       text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   timestamp   \n" +
				"  x mock_boolean_field                 boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"  x mock_boolean_field   true               false\n" +
				"  x mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_sqlite_equal_verbosity2",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsEqual1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences: none\n",
		},
		{
			"postgres_sqlite_left_table_absent_verbosity2",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsLeftTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_sqlite_right_table_absent_verbosity2",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsRightTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_sqlite_schema_different_verbosity2",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsSchema1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"postgres_sqlite_differences_verbosity2",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsMixed2Tables,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text2_field       text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   timestamp   \n" +
				"  x mock_boolean_field                 boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"  x mock_boolean_field   true               false\n" +
				"  x mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_postgres_equal_verbosity2",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsEqual1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"  = mock_boolean_field     boolean                       boolean\n" +
				"  = mock_id_field          text primary key              text primary key\n" +
				"  = mock_text_field        text                          text\n" +
				"  = mock_timestamp_field   timestamp without time zone   timestamp without time zone\n\n" +
				"  data differences: none\n",
		},
		{
			"postgres_postgres_left_table_absent_verbosity2",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_postgres_right_table_absent_verbosity2",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsRightTableAbsent1Table,
			2,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_postgres_schema_different_verbosity2",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsSchema1Table,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1          Database2\n" +
				"  = mock_boolean_field     boolean            boolean\n" +
				"  = mock_id_field          text primary key   text primary key\n" +
				"  = mock_text_field        text               text\n" +
				"  x mock_timestamp_field   boolean            timestamp without time zone\n\n",
		},
		{
			"postgres_postgres_differences_verbosity2",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsMixed2Tables,
			2,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"  = mock_id_field          text primary key              text primary key\n" +
				"  = mock_text2_field       text                          text\n" +
				"  = mock_text_field        text                          text\n" +
				"  x mock_timestamp_field   timestamp without time zone   \n" +
				"  x mock_boolean_field                                   boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"  = mock_boolean_field     boolean                       boolean\n" +
				"  = mock_id_field          text primary key              text primary key\n" +
				"  = mock_text_field        text                          text\n" +
				"  = mock_timestamp_field   timestamp without time zone   timestamp without time zone\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                Database1          Database2\n" +
				"  x mock_boolean_field   true               false\n" +
				"  x mock_text_field      mock_text_value1   mock_text_value1_diff\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		// verbosity 3
		{
			"sqlite_sqlite_equal_verbosity3",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsEqual1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"  = mock_boolean_field     boolean                                        boolean\n" +
				"  = mock_id_field          text primary key                               text primary key\n" +
				"  = mock_text_field        text not null unique                           text not null unique\n" +
				"  = mock_timestamp_field   timestamp not null default current_timestamp   timestamp not null default current_timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  = mock_text_field        mock_text_value1       mock_text_value1\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n",
		},
		{
			"sqlite_sqlite_left_table_absent_verbosity3",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsLeftTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_sqlite_right_table_absent_verbosity3",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsRightTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_sqlite_schema_different_verbosity3",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsSchema1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     boolean                boolean\n" +
				"  = mock_id_field          text primary key       text primary key\n" +
				"  x mock_text_field        text not null unique   text not null\n" +
				"  x mock_timestamp_field   boolean                timestamp not null default current_timestamp\n\n",
		},
		{
			"sqlite_sqlite_differences_verbosity3",
			mockSqliteDB1,
			mockSqliteDB2,
			nil,
			dbDiffsMixed2Tables,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"  = mock_id_field          text primary key                               text primary key\n" +
				"  = mock_text2_field       text not null                                  text not null\n" +
				"  = mock_text_field        text not null unique                           text not null unique\n" +
				"  x mock_timestamp_field   timestamp not null default current_timestamp   \n" +
				"  x mock_boolean_field                                                    boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                                      Database2\n" +
				"  = mock_boolean_field     boolean                                        boolean\n" +
				"  = mock_id_field          text primary key                               text primary key\n" +
				"  = mock_text_field        text not null unique                           text not null unique\n" +
				"  = mock_timestamp_field   timestamp not null default current_timestamp   timestamp not null default current_timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   false\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  x mock_text_field        mock_text_value1       mock_text_value1_diff\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"sqlite_postgres_equal_verbosity3",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsEqual1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  = mock_text_field        mock_text_value1       mock_text_value1\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n",
		},
		{
			"sqlite_postgres_left_table_absent_verbosity3",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"sqlite_postgres_right_table_absent_verbosity3",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsRightTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"sqlite_postgres_schema_different_verbosity3",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsSchema1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"sqlite_postgres_differences_verbosity3",
			mockSqliteDB1,
			mockPostgresDB2,
			testContainers{postgres2Container},
			dbDiffsMixed2Tables,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text2_field       text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   timestamp   \n" +
				"  x mock_boolean_field                 boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   false\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  x mock_text_field        mock_text_value1       mock_text_value1_diff\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_sqlite_equal_verbosity3",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsEqual1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  = mock_text_field        mock_text_value1       mock_text_value1\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n",
		},
		{
			"postgres_sqlite_left_table_absent_verbosity3",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsLeftTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_sqlite_right_table_absent_verbosity3",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsRightTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_sqlite_schema_different_verbosity3",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsSchema1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   boolean     timestamp\n\n",
		},
		{
			"postgres_sqlite_differences_verbosity3",
			mockPostgresDB1,
			mockSqliteDB2,
			testContainers{postgres1Container},
			dbDiffsMixed2Tables,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text2_field       text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  x mock_timestamp_field   timestamp   \n" +
				"  x mock_boolean_field                 boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1   Database2\n" +
				"  = mock_boolean_field     boolean     boolean\n" +
				"  = mock_id_field          text        text\n" +
				"  = mock_text_field        text        text\n" +
				"  = mock_timestamp_field   timestamp   timestamp\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   false\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  x mock_text_field        mock_text_value1       mock_text_value1_diff\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
		{
			"postgres_postgres_equal_verbosity3",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsEqual1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"  = mock_boolean_field     boolean                       boolean\n" +
				"  = mock_id_field          text primary key              text primary key\n" +
				"  = mock_text_field        text                          text\n" +
				"  = mock_timestamp_field   timestamp without time zone   timestamp without time zone\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  = mock_text_field        mock_text_value1       mock_text_value1\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n",
		},
		{
			"postgres_postgres_left_table_absent_verbosity3",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsLeftTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database1\n",
		},
		{
			"postgres_postgres_right_table_absent_verbosity3",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsRightTableAbsent1Table,
			3,
			"Table 'table0' does not exist in database2\n",
		},
		{
			"postgres_postgres_schema_different_verbosity3",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsSchema1Table,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1          Database2\n" +
				"  = mock_boolean_field     boolean            boolean\n" +
				"  = mock_id_field          text primary key   text primary key\n" +
				"  = mock_text_field        text               text\n" +
				"  x mock_timestamp_field   boolean            timestamp without time zone\n\n",
		},
		{
			"postgres_postgres_differences_verbosity3",
			mockPostgresDB1,
			mockPostgresDB2,
			testContainers{postgres1Container, postgres2Container},
			dbDiffsMixed2Tables,
			3,
			"Table table0:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"  = mock_id_field          text primary key              text primary key\n" +
				"  = mock_text2_field       text                          text\n" +
				"  = mock_text_field        text                          text\n" +
				"  x mock_timestamp_field   timestamp without time zone   \n" +
				"  x mock_boolean_field                                   boolean\n\n" +
				"Table table1:\n" +
				"  schema differences:\n" +
				"    Field                  Database1                     Database2\n" +
				"  = mock_boolean_field     boolean                       boolean\n" +
				"  = mock_id_field          text primary key              text primary key\n" +
				"  = mock_text_field        text                          text\n" +
				"  = mock_timestamp_field   timestamp without time zone   timestamp without time zone\n\n" +
				"  data differences:\n" +
				"  'mock_id_field'='id0':\n" +
				"    Field                  Database1              Database2\n" +
				"  = mock_boolean_field     true                   true\n" +
				"  = mock_id_field          id0                    id0\n" +
				"  = mock_text_field        mock_text_value0       mock_text_value0\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id1':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   false\n" +
				"  = mock_id_field          id1                    id1\n" +
				"  x mock_text_field        mock_text_value1       mock_text_value1_diff\n" +
				"  = mock_timestamp_field   2022-12-01T21:00:01Z   2022-12-01T21:00:01Z\n\n" +
				"  'mock_id_field'='id2':\n" +
				"    Field                  Database1              Database2\n" +
				"  x mock_boolean_field     true                   \n" +
				"  x mock_id_field          id2                    \n" +
				"  x mock_text_field        mock_text_value2       \n" +
				"  x mock_timestamp_field   2022-12-01T21:00:01Z   \n\n",
		},
	}

	postgresContainers := testContainers{postgres1Container, postgres2Container}

	setUp(t, postgresContainers)
	defer tearDown(t, postgresContainers)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetEnv(t, postgresContainers)

			generateDatabases(test.dbDifferences, &test.db1, &test.db2)
			require.NoError(t, test.db1.initialize())
			require.NoError(t, test.db2.initialize())

			done := captureOutput()
			c := comparer.NewDatabaseComparer(test.db1.mockInputString(), test.db2.mockInputString(), test.verbosity)
			c.Compare(context.Background())
			capturedOutput, err := done()
			require.NoError(t, err)

			require.ElementsMatch(t, strings.Split(capturedOutput, "\n"), strings.Split(test.expectedOutput, "\n"))
		})
	}
}
