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

type PoolOption func(opts *poolOptions)

func WithMigration(schemaVersion uint) PoolOption {
	return func(opts *poolOptions) {
		opts.migrateToVersion = &schemaVersion
	}
}

type poolOptions struct {
	migrateToVersion *uint
}

func OpenPool(ctx context.Context, url string, opts ...PoolOption) (*pgxpool.Pool, error) {
	var options poolOptions
	for _, opt := range opts {
		opt(&options)
	}

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

	if options.migrateToVersion != nil {
		if err = migrations.MigrateDatabase(db, *options.migrateToVersion); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return db, nil
}
