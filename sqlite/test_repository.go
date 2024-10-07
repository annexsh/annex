package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/annexsh/annex/test"
)

type testRepository struct {
	db *DB
	*ContextReader
	*ContextWriter
	*TestSuiteReader
	*TestSuiteWriter
	*TestReader
	*TestWriter
	*TestExecutionReader
	*TestExecutionWriter
	*CaseExecutionReader
	*CaseExecutionWriter
	*LogReader
	*LogWriter
}

func NewTestRepository(db *DB) test.Repository {
	return &testRepository{
		db:                  db,
		ContextReader:       NewContextReader(db),
		ContextWriter:       NewContextWriter(db),
		TestSuiteReader:     NewTestSuiteReader(db),
		TestSuiteWriter:     NewTestSuiteWriter(db),
		TestReader:          NewTestReader(db),
		TestWriter:          NewTestWriter(db),
		TestExecutionReader: NewTestExecutionReader(db),
		TestExecutionWriter: NewTestExecutionWriter(db),
		CaseExecutionReader: NewCaseExecutionReader(db),
		CaseExecutionWriter: NewCaseExecutionWriter(db),
		LogReader:           NewLogReader(db),
		LogWriter:           NewLogWriter(db),
	}
}

func (t *testRepository) WithTx(ctx context.Context) (test.Repository, test.Tx, error) {
	db, tx, err := t.db.WithTx(ctx)
	if err != nil {
		return nil, nil, err
	}

	return NewTestRepository(db), &txWrapper{base: tx}, err
}

func (t *testRepository) ExecuteTx(ctx context.Context, query func(repo test.Repository) error) error {
	repo, tx, err := t.WithTx(ctx)
	if err != nil {
		return err
	}

	if err = query(repo); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, fmt.Errorf("rollback error: %w", rbErr))
		}
		return err
	}

	return tx.Commit(ctx)
}

type txWrapper struct {
	base *sql.Tx
}

func (t *txWrapper) Commit(_ context.Context) error {
	return t.base.Commit()
}

func (t *txWrapper) Rollback(_ context.Context) error {
	return t.base.Rollback()
}
