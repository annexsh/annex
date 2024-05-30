package migrations

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed *.sql
var fs embed.FS

func MigrateDatabase(pool *pgxpool.Pool, version uint) error {
	sd, err := iofs.New(fs, ".")
	if err != nil {
		return err
	}
	defer sd.Close()

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	driver, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return err
	}
	defer driver.Close()

	m, err := migrate.NewWithInstance("iofs", sd, "postgres", driver)
	if err != nil {
		return err
	}

	if err = m.Migrate(version); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
