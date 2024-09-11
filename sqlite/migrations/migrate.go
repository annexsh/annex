package migrations

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var fs embed.FS

func MigrateDatabase(db *sql.DB) error {
	sd, err := iofs.New(fs, ".")
	if err != nil {
		return err
	}
	defer sd.Close()

	driver, err := sqlite.WithInstance(db, new(sqlite.Config))
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sd, "sqlite", driver)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
