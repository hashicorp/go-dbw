package dbw

import (
	"context"
	"fmt"
	"reflect"
)

// Delete a resource in the db with options: WithWhere, WithDebug, WithTable,
// and WithVersion. WithWhere and WithVersion allows specifying a additional
// constraints on the operation in addition to the PKs. Delete returns the
// number of rows deleted and any errors.
func (rw *RW) Delete(ctx context.Context, i interface{}, opt ...Option) (int, error) {
	const op = "dbw.Delete"
	if rw.underlying == nil {
		return noRowsAffected, fmt.Errorf("%s: missing underlying db: %w", op, ErrInvalidParameter)
	}
	if isNil(i) {
		return noRowsAffected, fmt.Errorf("%s: missing interface: %w", op, ErrInvalidParameter)
	}
	if err := raiseErrorOnHooks(i); err != nil {
		return noRowsAffected, fmt.Errorf("%s: %w", op, err)
	}
	opts := GetOpts(opt...)

	mDb := rw.underlying.wrapped.Model(i)
	err := mDb.Statement.Parse(i)
	if err == nil && mDb.Statement.Schema == nil {
		return noRowsAffected, fmt.Errorf("%s: (internal error) unable to parse stmt: %w", op, ErrUnknown)
	}
	reflectValue := reflect.Indirect(reflect.ValueOf(i))
	for _, pf := range mDb.Statement.Schema.PrimaryFields {
		if _, isZero := pf.ValueOf(reflectValue); isZero {
			return noRowsAffected, fmt.Errorf("%s: primary key %s is not set: %w", op, pf.Name, ErrInvalidParameter)
		}
	}
	if opts.WithBeforeWrite != nil {
		if err := opts.WithBeforeWrite(i); err != nil {
			return noRowsAffected, fmt.Errorf("%s: error before write: %w", op, err)
		}
	}
	db := rw.underlying.wrapped
	if opts.WithVersion != nil || opts.WithWhereClause != "" {
		where, args, err := rw.whereClausesFromOpts(ctx, i, opts)
		if err != nil {
			return noRowsAffected, fmt.Errorf("%s: %w", op, err)
		}
		db = db.Where(where, args...)
	}
	if opts.WithDebug {
		db = db.Debug()
	}
	if opts.WithTable != "" {
		db = db.Table(opts.WithTable)
	}
	db = db.Delete(i)
	if db.Error != nil {
		return noRowsAffected, fmt.Errorf("%s: %w", op, db.Error)
	}
	rowsDeleted := int(db.RowsAffected)
	if rowsDeleted > 0 && opts.WithAfterWrite != nil {
		if err := opts.WithAfterWrite(i, rowsDeleted); err != nil {
			return rowsDeleted, fmt.Errorf("%s: error after write: %w", op, err)
		}
	}
	return rowsDeleted, nil
}

// DeleteItems will delete multiple items of the same type. Options supported:
// WithDebug, WithTable
func (rw *RW) DeleteItems(ctx context.Context, deleteItems []interface{}, opt ...Option) (int, error) {
	const op = "dbw.DeleteItems"
	if rw.underlying == nil {
		return noRowsAffected, fmt.Errorf("%s: missing underlying db: %w", op, ErrInvalidParameter)
	}
	if len(deleteItems) == 0 {
		return noRowsAffected, fmt.Errorf("%s: no interfaces to delete: %w", op, ErrInvalidParameter)
	}
	if err := raiseErrorOnHooks(deleteItems); err != nil {
		return noRowsAffected, fmt.Errorf("%s: %w", op, err)
	}
	opts := GetOpts(opt...)
	if opts.WithLookup {
		return noRowsAffected, fmt.Errorf("%s: with lookup not a supported option: %w", op, ErrInvalidParameter)
	}
	// verify that createItems are all the same type.
	var foundType reflect.Type
	for i, v := range deleteItems {
		if i == 0 {
			foundType = reflect.TypeOf(v)
		}
		currentType := reflect.TypeOf(v)
		if foundType != currentType {
			return noRowsAffected, fmt.Errorf("%s: items contain disparate types.  item %d is not a %s: %w", op, i, foundType.Name(), ErrInvalidParameter)
		}
	}
	if opts.WithBeforeWrite != nil {
		if err := opts.WithBeforeWrite(deleteItems); err != nil {
			return noRowsAffected, fmt.Errorf("%s: error before write: %w", op, err)
		}
	}
	rowsDeleted := 0
	for _, item := range deleteItems {
		cnt, err := rw.Delete(ctx, item,
			WithDebug(opts.WithDebug),
			WithTable(opts.WithTable),
		)
		rowsDeleted += cnt
		if err != nil {
			return rowsDeleted, fmt.Errorf("%s: %w", op, err)
		}
	}
	if rowsDeleted > 0 && opts.WithAfterWrite != nil {
		if err := opts.WithAfterWrite(deleteItems, int(rowsDeleted)); err != nil {
			return rowsDeleted, fmt.Errorf("%s: error after write: %w", op, err)
		}
	}
	return rowsDeleted, nil
}
