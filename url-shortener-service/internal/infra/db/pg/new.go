package pg

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	pkgerrors "github.com/pkg/errors"
)

// DatabaseName indicates name of applying database. It's `postgres` in this case
const DatabaseName = "postgres"

// Connect connects to database
func Connect(dbURL string) (*sql.DB, error) {
	conn, err := sql.Open(DatabaseName, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	if err = conn.Ping(); err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	return conn, nil
}
