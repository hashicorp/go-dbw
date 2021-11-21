package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Query will run the raw query and return the *sql.Rows results. Query will
// operate within the context of any ongoing transaction for the Reader.  The
// caller must close the returned *sql.Rows. Query can/should be used in
// combination with ScanRows.
func (rw *RW) Query(ctx context.Context, sql string, values []interface{}, _ ...Option) (*sql.Rows, error) {
	const op = "db.Query"
	if rw.underlying == nil {
		return nil, fmt.Errorf("%s: missing underlying db: %w", op, ErrInternal)
	}
	if sql == "" {
		return nil, fmt.Errorf("%s: missing sql: %w", op, ErrInvalidParameter)
	}
	gormDb := rw.underlying.Raw(sql, values...)
	if gormDb.Error != nil {
		return nil, fmt.Errorf("%s: %w", op, gormDb.Error)
	}
	return gormDb.Rows()
}

// Scan rows will scan the rows into the interface
func (rw *RW) ScanRows(rows *sql.Rows, result interface{}) error {
	const op = "db.ScanRows"
	if rw.underlying == nil {
		return fmt.Errorf("%s: missing underlying db: %w", op, ErrInternal)
	}
	if isNil(result) {
		return fmt.Errorf("%s: missing result: %w", op, ErrInvalidParameter)
	}
	return rw.underlying.ScanRows(rows, result)
}
