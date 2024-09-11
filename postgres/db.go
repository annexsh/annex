package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/annexsh/annex/postgres/sqlc"
)

type DBTX interface {
	sqlc.DBTX
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type DB struct {
	*sqlc.Queries
	beginTx func(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func NewDB(dbtx DBTX) *DB {
	return &DB{
		Queries: sqlc.New(dbtx),
		beginTx: dbtx.BeginTx,
	}
}

func (d *DB) WithTx(ctx context.Context) (*DB, pgx.Tx, error) {
	tx, err := d.beginTx(ctx, pgx.TxOptions{})
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
