package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/annexsh/annex/postgres/migrations"
)

const (
	dbName          = "annex"
	pgRetryInterval = 2 * time.Second
	pgMaxRetries    = 5
)

type OpenOption func(opts *openOptions)

func WithMigration() OpenOption {
	return func(opts *openOptions) {
		opts.migrateUp = true
	}
}

type openOptions struct {
	migrateUp bool
}

func OpenPool(ctx context.Context, username string, password string, hostPort string, opts ...OpenOption) (*pgxpool.Pool, error) {
	var options openOptions
	for _, opt := range opts {
		opt(&options)
	}

	postgresURL := fmt.Sprintf("postgres://%s:%s@%s/postgres", username, password, hostPort)
	pool, err := newPool(ctx, postgresURL)
	if err != nil {
		return nil, err
	}

	if _, err = pool.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code != pgerrcode.DuplicateDatabase {
				return nil, err
			}
		}
	}

	pool.Close()

	annexURL := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, hostPort, dbName)
	pool, err = newPool(ctx, annexURL)
	if err != nil {
		return nil, err
	}

	if options.migrateUp {
		if err = migrations.MigrateDatabase(pool); err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return pool, nil
}

func newPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	operation := func() error {
		return pool.Ping(ctx)
	}

	bo := backoff.WithMaxRetries(backoff.NewConstantBackOff(pgRetryInterval), pgMaxRetries)
	if err = backoff.Retry(operation, bo); err != nil {
		return nil, fmt.Errorf("database connection unhealthy: %w", err)
	}

	return pool, nil
}
