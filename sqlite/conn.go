package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/annexsh/annex/sqlite/migrations"
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

func Open(opts ...OpenOption) (*sql.DB, error) {
	var options openOptions
	for _, opt := range opts {
		opt(&options)
	}

	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database connection: %w", err)
	}
	if _, err = db.Exec("PRAGMA foreign_keys = on;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if options.migrateUp {
		if err = migrations.MigrateDatabase(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return db, nil
}
