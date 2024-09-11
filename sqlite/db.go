package sqlite

import (
	"context"
	"database/sql"

	"github.com/annexsh/annex/sqlite/sqlc"
)

type DBTX interface {
	sqlc.DBTX
	BeginTx(ctx context.Context, txOptions *sql.TxOptions) (*sql.Tx, error)
}

type DB struct {
	*sqlc.Queries
	beginTx func(ctx context.Context, txOptions *sql.TxOptions) (*sql.Tx, error)
}

func NewDB(dbtx DBTX) *DB {
	return &DB{
		Queries: sqlc.New(dbtx),
		beginTx: dbtx.BeginTx,
	}
}

func (d *DB) WithTx(ctx context.Context) (*DB, *sql.Tx, error) {
	tx, err := d.beginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, nil, err
	}

	queries := d.Queries.WithTx(tx)

	newDB := &DB{
		Queries: queries,
		beginTx: d.beginTx,
	}

	return newDB, tx, nil
}
