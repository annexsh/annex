package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/annexsh/annex/postgres/migrations"
)

const (
	pgRetryInterval = 2 * time.Second
	pgMaxRetries    = 5
)

func OpenPool(ctx context.Context, url string, schemaVersion uint) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	operation := func() error {
		return db.Ping(ctx)
	}

	bo := backoff.WithMaxRetries(backoff.NewConstantBackOff(pgRetryInterval), pgMaxRetries)
	if err = backoff.Retry(operation, bo); err != nil {
		return nil, fmt.Errorf("database connection unhealthy: %w", err)
	}

	if err = migrations.MigrateDatabase(db, schemaVersion); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
