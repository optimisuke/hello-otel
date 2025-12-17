package db

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	// Register pgx stdlib driver for database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// New creates a sqlx.DB backed by pgx stdlib driver.
func New(ctx context.Context, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Keep timeouts modest to fail fast on bad connections.
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	// Verify connectivity early.
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
