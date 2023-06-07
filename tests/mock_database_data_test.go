package tests

import (
	"context"
	"fmt"

	"github.com/ygrebnov/testutils/docker"
	"github.com/ygrebnov/testutils/presets"

	"github.com/ygrebnov/dbdiff/dbdiff"
	"github.com/ygrebnov/dbdiff/models"
)

var (
	postgresUser         = "postgres"
	postgresPassword     = "postgres"
	postgresPort1        = 5433
	postgresPort2        = 5434
	exposedPostgresPort1 = []string{fmt.Sprintf("%d:5432", postgresPort1)}
	exposedPostgresPort2 = []string{fmt.Sprintf("%d:5432", postgresPort2)}
	postgres1Container   = presets.NewCustomizedPostgresqlContainer(docker.Options{Name: "postgres1", ExposedPorts: exposedPostgresPort1})
	postgres2Container   = presets.NewCustomizedPostgresqlContainer(docker.Options{Name: "postgres2", ExposedPorts: exposedPostgresPort2})
	mockPostgresDB1      = newMockDatabase(dbTypePostgresql, &postgres1Container, &postgresPort1, "")
	mockPostgresDB2      = newMockDatabase(dbTypePostgresql, &postgres2Container, &postgresPort2, "")
	mockSqliteDB1        = newMockDatabase(dbTypeSqlite, nil, nil, "test/db1")
	mockSqliteDB2        = newMockDatabase(dbTypeSqlite, nil, nil, "test/db2")
	mockIDField          = models.Field{Name: "mock_id_field", FieldType: "text", PrimaryKey: true, Attrs: "primary key"}
	mockTextField        = models.Field{Name: "mock_text_field", FieldType: "text", PrimaryKey: false, Attrs: "not null"}
	mockText2Field       = models.Field{Name: "mock_text2_field", FieldType: "text", PrimaryKey: false, Attrs: "not null"}
	mockBooleanField     = models.Field{Name: "mock_boolean_field", FieldType: "boolean", PrimaryKey: false, Attrs: "not null"}
	mockTimestampField   = models.Field{Name: "mock_timestamp_field", FieldType: "timestamp", PrimaryKey: false, Attrs: "not null default current_timestamp"}
	mockFields           = []*models.Field{&mockIDField, &mockTextField, &mockBooleanField, &mockTimestampField}
	mockFields5          = []*models.Field{&mockIDField, &mockTextField, &mockBooleanField, &mockTimestampField, &mockText2Field}
	mockV0Ctx            = context.Background()
	mockV1Ctx            = context.WithValue(context.Background(), dbdiff.VerboseContextKey, true)
	mockV2Ctx            = context.WithValue(context.Background(), dbdiff.VVerboseContextKey, true)
	mockV3Ctx            = context.WithValue(context.Background(), dbdiff.VVVerboseContextKey, true)
	tbDiffsEqual         = []tableDifferences{
		{
			2,
			exists{inD1: true, inD2: true},
			mockFields,
			map[int]exists{
				0: {true, true},
				1: {true, true},
				2: {true, true},
				3: {true, true},
			},
			map[int]schemaDifferences{
				0: {"text", "text"},
				1: {"text", "text"},
				2: {"boolean", "boolean"},
				3: {"timestamp", "timestamp"},
			},
			map[int]schemaDifferences{
				0: {"primary key", "primary key"},
				1: {"not null unique", "not null unique"},
				2: {"", ""},
				3: {"not null default current_timestamp", "not null default current_timestamp"},
			},
			map[int]dataDifferences{},
		},
	}
	dbDiffsEqual1Table = databaseDifferences{1, tbDiffsEqual}

	tbDiffsLeftTableAbsent = []tableDifferences{
		{
			2,
			exists{inD1: false, inD2: true},
			mockFields,
			map[int]exists{
				0: {true, true},
				1: {true, true},
				2: {true, true},
				3: {true, true},
			},
			map[int]schemaDifferences{
				0: {"text", "text"},
				1: {"text", "text"},
				2: {"boolean", "boolean"},
				3: {"timestamp", "timestamp"},
			},
			map[int]schemaDifferences{
				0: {"primary key", "primary key"},
				1: {"not null unique", "not null unique"},
				2: {"", ""},
				3: {"not null default current_timestamp", "not null default current_timestamp"},
			},
			map[int]dataDifferences{},
		},
	}
	dbDiffsLeftTableAbsent1Table = databaseDifferences{1, tbDiffsLeftTableAbsent}

	tbDiffsRightTableAbsent = []tableDifferences{
		{
			2,
			exists{inD1: true, inD2: false},
			mockFields,
			map[int]exists{
				0: {true, true},
				1: {true, true},
				2: {true, true},
				3: {true, true},
			},
			map[int]schemaDifferences{
				0: {"text", "text"},
				1: {"text", "text"},
				2: {"boolean", "boolean"},
				3: {"timestamp", "timestamp"},
			},
			map[int]schemaDifferences{
				0: {"primary key", "primary key"},
				1: {"not null unique", "not null unique"},
				2: {"", ""},
				3: {"not null default current_timestamp", "not null default current_timestamp"},
			},
			map[int]dataDifferences{},
		},
	}
	dbDiffsRightTableAbsent1Table = databaseDifferences{1, tbDiffsRightTableAbsent}

	tbDiffsSchema = []tableDifferences{
		{
			2,
			exists{inD1: true, inD2: true},
			mockFields,
			map[int]exists{
				0: {true, true},
				1: {true, true},
				2: {true, true},
				3: {true, true},
			},
			map[int]schemaDifferences{
				0: {"text", "text"},
				1: {"text", "text"},
				2: {"boolean", "boolean"},
				3: {"boolean", "timestamp"},
			},
			map[int]schemaDifferences{
				0: {"primary key", "primary key"},
				1: {"not null unique", "not null"},
				2: {"", ""},
				3: {"", "not null default current_timestamp"},
			},
			map[int]dataDifferences{
				0: {
					"'id0', 'mock_text_value0', 'TRUE', 'FALSE'",
					"'id0', 'mock_text_value0', 'TRUE', '2022-12-01 21:00:01'",
				},
				1: {
					"'id1', 'mock_text_value1', 'TRUE', 'FALSE'",
					"'id1', 'mock_text_value1', 'TRUE', '2022-12-01 21:00:02'",
				},
			},
		},
	}
	dbDiffsSchema1Table = databaseDifferences{1, tbDiffsSchema}

	tbDiffsMixed = []tableDifferences{
		{
			2,
			exists{inD1: true, inD2: true},
			mockFields5,
			map[int]exists{
				0: {true, true},
				1: {true, true},
				2: {false, true},
				3: {true, false},
				4: {true, true},
			},
			map[int]schemaDifferences{
				0: {"text", "text"},
				1: {"text", "text"},
				2: {"boolean", "boolean"},
				3: {"timestamp", "timestamp"},
				4: {"text", "text"},
			},
			map[int]schemaDifferences{
				0: {"primary key", "primary key"},
				1: {"not null unique", "not null unique"},
				2: {"", ""},
				3: {"not null default current_timestamp", "not null default current_timestamp"},
				4: {"not null", "not null"},
			},
			map[int]dataDifferences{
				0: {
					"'id0', 'mock_text_value0', '2022-12-01 21:00:01', 'mock_text_value0'",
					"'id0', 'mock_text_value0', 'TRUE', 'mock_text_value0_diff'",
				},
				1: {
					"'id1', 'mock_text_value1', '2022-12-01 21:00:01', 'mock_text_value1'",
					"'id1', 'mock_text_value1', 'TRUE', 'mock_text_value1'",
				},
			},
		},
		{
			3,
			exists{inD1: true, inD2: true},
			mockFields,
			map[int]exists{
				0: {true, true},
				1: {true, true},
				2: {true, true},
				3: {true, true},
			},
			map[int]schemaDifferences{
				0: {"text", "text"},
				1: {"text", "text"},
				2: {"boolean", "boolean"},
				3: {"timestamp", "timestamp"},
			},
			map[int]schemaDifferences{
				0: {"primary key", "primary key"},
				1: {"not null unique", "not null unique"},
				2: {"", ""},
				3: {"not null default current_timestamp", "not null default current_timestamp"},
			},
			map[int]dataDifferences{
				0: {
					"'id0', 'mock_text_value0', 'TRUE', '2022-12-01 21:00:01'",
					"'id0', 'mock_text_value0', 'TRUE', '2022-12-01 21:00:01'",
				},
				1: {
					"'id1', 'mock_text_value1', 'TRUE', '2022-12-01 21:00:01'",
					"'id1', 'mock_text_value1_diff', 'FALSE', '2022-12-01 21:00:01'",
				},
				2: {
					"'id2', 'mock_text_value2', 'TRUE', '2022-12-01 21:00:01'",
					"",
				},
			},
		},
	}
	dbDiffsMixed2Tables = databaseDifferences{2, tbDiffsMixed}
)
