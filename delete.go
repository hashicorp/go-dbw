package dbw

import (
	"context"
	"fmt"
	"reflect"
)

// Delete a resource in the db with options: WithWhere, WithDebug and
// WithVersion. WithWhere and WithVersion allows specifying a additional
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
	if opts.withBeforeWrite != nil {
		if err := opts.withBeforeWrite(i); err != nil {
			return noRowsAffected, fmt.Errorf("%s: error before write: %w", op, err)
		}
	}
	db := rw.underlying.wrapped
	if opts.withWhereClause != "" {
		db = db.Where(opts.withWhereClause, opts.withWhereClauseArgs...)
	}
	if opts.withDebug {
		db = db.Debug()
	}
	db = db.Delete(i)
	if db.Error != nil {
		return noRowsAffected, fmt.Errorf("%s: %w", op, db.Error)
	}
	rowsDeleted := int(db.RowsAffected)
	if rowsDeleted > 0 && opts.withAfterWrite != nil {
		if err := opts.withAfterWrite(i, rowsDeleted); err != nil {
			return rowsDeleted, fmt.Errorf("%s: error after write: %w", op, err)
		}
	}
	return rowsDeleted, nil
}

// DeleteItems will delete multiple items of the same type. Options supported: WithDebug
func (rw *RW) DeleteItems(ctx context.Context, deleteItems []interface{}, opt ...Option) (int, error) {
	const op = "dbw.DeleteItems"
	if rw.underlying == nil {
		return noRowsAffected, fmt.Errorf("%s: missing underlying db: %w", op, ErrInvalidParameter)
	}
	if len(deleteItems) == 0 {
		return noRowsAffected, fmt.Errorf("%s: no interfaces to delete: %w", op, ErrInvalidParameter)
	}
	opts := GetOpts(opt...)
	if opts.withLookup {
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
	if opts.withBeforeWrite != nil {
		if err := opts.withBeforeWrite(deleteItems); err != nil {
			return noRowsAffected, fmt.Errorf("%s: error before write: %w", op, err)
		}
	}
	rowsDeleted := 0
	for _, item := range deleteItems {
		cnt, err := rw.Delete(ctx, item,
			WithDebug(opts.withDebug),
		)
		rowsDeleted += cnt
		if err != nil {
			return rowsDeleted, fmt.Errorf("%s: %w", op, err)
		}
	}
	if rowsDeleted > 0 && opts.withAfterWrite != nil {
		if err := opts.withAfterWrite(deleteItems, int(rowsDeleted)); err != nil {
			return rowsDeleted, fmt.Errorf("%s: error after write: %w", op, err)
		}
	}
	return rowsDeleted, nil
}
