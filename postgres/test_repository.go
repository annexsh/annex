package postgres

import (
	"context"
	"fmt"

	"github.com/annexsh/annex/test"
)

type testRepository struct {
	db *DB
	*ContextReader
	*ContextWriter
	*GroupReader
	*GroupWriter
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
		GroupReader:         NewGroupReader(db),
		GroupWriter:         NewGroupWriter(db),
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

	return NewTestRepository(db), tx, err
}

func (t *testRepository) ExecuteTx(ctx context.Context, query func(repo test.Repository) error) (err error) {
	repo, tx, err := t.WithTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if rErr := tx.Rollback(ctx); rErr != nil {
				panic(fmt.Errorf("panic: %v; failed to rollback transaction: %w", r, rErr))
			}
			panic(r)
		}

		if err != nil {
			if rErr := tx.Rollback(ctx); rErr != nil {
				err = fmt.Errorf("%w; failed to rollback transaction: %w", err, rErr)
			}
			return
		}

		if cErr := tx.Commit(ctx); cErr != nil {
			err = fmt.Errorf("failed to commit transaction: %w", cErr)
		}
	}()

	return query(repo)
}
