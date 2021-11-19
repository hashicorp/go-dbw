package db

import (
	"context"
	"fmt"
)

const (
	NoRowsAffected = 0

	// DefaultLimit is the default for results for boundary
	DefaultLimit = 10000
)

// RW uses a gorm DB connection for read/write
type RW struct {
	underlying *DB
}

func New(underlying *DB) *RW {
	return &RW{underlying: underlying}
}

// Exec will execute the sql with the values as parameters. The int returned
// is the number of rows affected by the sql. No options are currently
// supported.
func (rw *RW) Exec(ctx context.Context, sql string, values []interface{}, _ ...Option) (int, error) {
	const op = "db.Exec"
	if sql == "" {
		return NoRowsAffected, fmt.Errorf("%s: missing sql: %w", op, ErrInvalidParameter)
	}
	gormDb := rw.underlying.Exec(sql, values...)
	if gormDb.Error != nil {
		return NoRowsAffected, fmt.Errorf("%s: %w", op, gormDb.Error)
	}
	return int(gormDb.RowsAffected), nil
}
