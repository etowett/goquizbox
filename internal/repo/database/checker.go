package database

import (
	"context"
	"errors"
	"fmt"

	"goquizbox/pkg/database"

	pgx "github.com/jackc/pgx/v4"
)

const (
	getDatabaseVersionSQL = `select version()`
	getDatabaseOneSQL     = `select 1`
)

// CheckerDB is a handle to database operations for testing db liveness
type CheckerDB struct {
	db *database.DB
}

// New creates a new CheckerDB that wraps a raw database handle.
func NewCheckerDB(db *database.DB) *CheckerDB {
	return &CheckerDB{
		db: db,
	}
}

// SelectOne tests the database with select 1
func (c *CheckerDB) SelectOne(ctx context.Context) (*int32, error) {
	var val *int32

	if err := c.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getDatabaseOneSQL)

		var err error
		val, err = scanOneCheck(row)
		if err != nil {
			return fmt.Errorf("failed to parse scan one: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("select one: %w", err)
	}

	return val, nil
}

// SelectVersion tests the database with select version
func (c *CheckerDB) SelectVersion(ctx context.Context) (*string, error) {
	var val string

	if err := c.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getDatabaseVersionSQL)

		var err error
		val, err = scanVersionCheck(row)
		if err != nil {
			return fmt.Errorf("failed to parse scan version: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("select version: %w", err)
	}

	return &val, nil
}

func scanOneCheck(row pgx.Row) (*int32, error) {
	var val int32
	if err := row.Scan(&val); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &val, nil
}

func scanVersionCheck(row pgx.Row) (string, error) {
	var val string
	if err := row.Scan(&val); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return val, nil
}
