package dbw

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// LookupByPublicId will lookup resource by its public_id or private_id, which
// must be unique. Options are ignored.
func (rw *RW) LookupById(ctx context.Context, resourceWithIder interface{}, _ ...Option) error {
	const op = "dbw.LookupById"
	if rw.underlying == nil {
		return fmt.Errorf("%s: missing underlying db: %w", op, ErrInvalidParameter)
	}
	if reflect.ValueOf(resourceWithIder).Kind() != reflect.Ptr {
		return fmt.Errorf("%s: interface parameter must to be a pointer: %w", op, ErrInvalidParameter)
	}
	where, keys, err := rw.primaryKeysWhere(ctx, resourceWithIder)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := rw.underlying.Where(where, keys...).First(resourceWithIder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%s: %w", op, ErrRecordNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// LookupByPublicId will lookup resource by its public_id, which must be unique.
// Options are ignored.
func (rw *RW) LookupByPublicId(ctx context.Context, resource ResourcePublicIder, opt ...Option) error {
	return rw.LookupById(ctx, resource, opt...)
}

func (rw *RW) lookupAfterWrite(ctx context.Context, i interface{}, opt ...Option) error {
	const op = "dbw.lookupAfterWrite"
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
