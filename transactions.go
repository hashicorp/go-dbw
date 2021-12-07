package dbw

import (
	"context"
	"fmt"
)

// Begin will start a transaction
func (rw *RW) Begin(ctx context.Context) (*RW, error) {
	const op = "dbw.Begin"
	newTx := rw.underlying.wrapped.WithContext(ctx)
	newTx = newTx.Begin()
	if newTx.Error != nil {
		return nil, fmt.Errorf("%s: %w", op, newTx.Error)
	}
	return New(
		&DB{wrapped: newTx},
	), nil
}

// Rollback will rollback the current transaction
func (rw *RW) Rollback(_ context.Context) error {
	const op = "dbw.Rollback"
	if err := rw.underlying.wrapped.Rollback().Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Commit will commit a transaction
func (rw *RW) Commit(_ context.Context) error {
	const op = "dbw.Commit"
	if err := rw.underlying.wrapped.Commit().Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
