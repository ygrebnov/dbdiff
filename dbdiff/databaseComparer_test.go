package dbdiff

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygrebnov/dbdiff/models"
)

func TestParse(t *testing.T) {
	osStat = func(name string) (any, error) { // mock os.Stat function
		return nil, nil
	}

	var tests = []struct {
		input          string
		expectedDBType models.DatabaseType
		expectedDBURI  string
	}{
		{
			"sqlite:pr.sql",
			sqlite,
			"pr.sql",
		},
		{
			"postgres:postgres://username:userpassword@hostname:port/dbname",
			postgresql,
			"postgres://username:userpassword@hostname:port/dbname",
		},
	}

	for _, test := range tests {
		var d models.Database
		t.Run(test.input, func(t *testing.T) {
			NewDatabaseComparer().parse(test.input, &d)
			require.Equal(t, test.expectedDBType, d.DBType)
			require.Equal(t, test.expectedDBURI, d.URI)
		})
	}
}
