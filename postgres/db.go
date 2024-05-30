package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/annexhq/annex/postgres/sqlc"
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

func (d *DB) WithTx(ctx context.Context) (pgx.Tx, sqlc.Querier, error) {
	tx, err := d.beginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}

	queries := d.Queries.WithTx(tx)
	return tx, queries, nil
}

func (d *DB) ExecuteTx(ctx context.Context, query func(querier sqlc.Querier) error) error {
	tx, err := d.beginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	withTx := d.Queries.WithTx(tx)
	if err = query(withTx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, fmt.Errorf("rollback error: %w", rbErr))
		}
		return err
	}

	return tx.Commit(ctx)
}
