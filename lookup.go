package db

import (
	"context"
	"fmt"
)

func (rw *RW) lookupAfterWrite(ctx context.Context, i interface{}, opt ...Option) error {
	const op = "db.lookupAfterWrite"
	opts := getOpts(opt...)
	withLookup := opts.withLookup

	if !withLookup {
		return nil
	}
	if err := rw.LookupById(ctx, i, opt...); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
