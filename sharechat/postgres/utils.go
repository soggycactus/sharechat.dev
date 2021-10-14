package postgres

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
)

func executeTransaction(ctx context.Context, db sqlx.DB, query string, args ...interface{}) error {
	var err error
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = multierror.Append(err, rollbackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = multierror.Append(err, rollbackErr)
		}
		return err
	}

	return nil
}
