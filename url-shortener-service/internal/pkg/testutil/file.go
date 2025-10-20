package testutil

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// LoadSQLFile load test sql data from a file
func LoadSQLFile(t *testing.T, db *sql.Tx, sqlFile string) {
	b, err := os.ReadFile(sqlFile)
	require.NoError(t, err)

	_, err = db.Exec(string(b))
	require.NoError(t, err)
}
