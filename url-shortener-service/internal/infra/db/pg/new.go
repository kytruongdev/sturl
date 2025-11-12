package pg

import (
	"context"
	"database/sql"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	pkgerrors "github.com/pkg/errors"
)

// DatabaseName indicates the name of the database driver.
const DatabaseName = "pgx"

// Connect establishes a connection to the PostgreSQL database with OpenTelemetry instrumentation.
func Connect(dbURL string) (*sql.DB, error) {
	// register instrumented driver for PostgreSQL
	driverName, err := otelsql.Register(DatabaseName, otelsql.WithAttributes())
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// open DB using the instrumented driver
	db, err := sql.Open(driverName, dbURL)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	return db, nil
}
