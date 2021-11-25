package dbw

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm/clause"
)

type OpType int

const (
	UnknownOp OpType = 0
	CreateOp  OpType = 1
	UpdateOp  OpType = 2
	DeleteOp  OpType = 3
)

// VetForWriter provides an interface that Create and Update can use to vet the
// resource before before writing it to the db.  For optType == UpdateOp,
// options WithFieldMaskPath and WithNullPaths are supported.  For optType ==
// CreateOp, no options are supported
type VetForWriter interface {
	VetForWrite(ctx context.Context, r Reader, opType OpType, opt ...Option) error
}

// Create an object in the db with options: WithDebug, WithLookup,
// WithReturnRowsAffected, OnConflict, WithBeforeWrite, WithAfterWrite,
// WithVersion, and WithWhere.
//
// OnConflict specifies alternative actions to take when an insert results in a
// unique constraint or exclusion constraint error. If WithVersion is used, then
// the update for on conflict will include the version number, which basically
// makes the update use optimistic locking and the update will only succeed if
// the existing rows version matches the WithVersion option.  Zero is not a
// valid value for the WithVersion option and will return an error. WithWhere
// allows specifying an additional constraint on the on conflict operation in
// addition to the on conflict target policy (columns or constraint).
func (rw *RW) Create(ctx context.Context, i interface{}, opt ...Option) error {
	const op = "dbw.Create"
	if rw.underlying == nil {
		return fmt.Errorf("%s: missing underlying db: %w", op, ErrInvalidParameter)
	}
	if isNil(i) {
		return fmt.Errorf("%s: missing interface: %w", op, ErrInvalidParameter)
	}
	opts := getOpts(opt...)

	// these fields should be nil, since they are not writeable and we want the
	// db to manage them
	setFieldsToNil(i, []string{"CreateTime", "UpdateTime"})

	if !opts.withSkipVetForWrite {
		if vetter, ok := i.(VetForWriter); ok {
			if err := vetter.VetForWrite(ctx, rw, CreateOp); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	db := rw.underlying.wrapped.WithContext(ctx)
	var onConflictDoNothing bool
	if opts.withOnConflict != nil {
		c := clause.OnConflict{}
		switch opts.withOnConflict.Target.(type) {
		case Constraint:
			c.OnConstraint = string(opts.withOnConflict.Target.(Constraint))
		case Columns:
			columns := make([]clause.Column, 0, len(opts.withOnConflict.Target.(Columns)))
			for _, name := range opts.withOnConflict.Target.(Columns) {
				columns = append(columns, clause.Column{Name: name})
			}
			c.Columns = columns
		default:
			return fmt.Errorf("%s: invalid conflict target %v: %w", op, reflect.TypeOf(opts.withOnConflict.Target), ErrInvalidParameter)
		}

		switch opts.withOnConflict.Action.(type) {
		case DoNothing:
			c.DoNothing = true
			onConflictDoNothing = true
		case UpdateAll:
			c.UpdateAll = true
		case []ColumnValue:
			updates := opts.withOnConflict.Action.([]ColumnValue)
			set := make(clause.Set, 0, len(updates))
			for _, s := range updates {
				// make sure it's not one of the std immutable columns
				if contains([]string{"createtime", "publicid"}, strings.ToLower(s.column)) {
					return fmt.Errorf("%s: cannot do update on conflict for column %s: %w", op, s.column, ErrInvalidParameter)
				}
				switch sv := s.value.(type) {
				case column:
					set = append(set, sv.toAssignment(s.column))
				case ExprValue:
					set = append(set, sv.toAssignment(s.column))
				default:
					set = append(set, rawAssignment(s.column, s.value))
				}
			}
			c.DoUpdates = set
		default:
			return fmt.Errorf("%s: invalid conflict action %v: %w", op, reflect.TypeOf(opts.withOnConflict.Action), ErrInvalidParameter)
		}
		if opts.WithVersion != nil || opts.withWhereClause != "" {
			where, args, err := rw.whereClausesFromOpts(ctx, i, opts)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			whereConditions := db.Statement.BuildCondition(where, args...)
			c.Where = clause.Where{Exprs: whereConditions}
		}
		db = db.Clauses(c)
	}
	if opts.withDebug {
		db = db.Debug()
	}

	if opts.withBeforeWrite != nil {
		if err := opts.withBeforeWrite(i); err != nil {
			return fmt.Errorf("%s: error before write: %w", op, err)
		}
	}
	tx := db.Create(i)
	if tx.Error != nil {
		return fmt.Errorf("%s: create failed: %w", op, tx.Error)
	}
	if opts.withRowsAffected != nil {
		*opts.withRowsAffected = tx.RowsAffected
	}
	if opts.withAfterWrite != nil {
		switch {
		case onConflictDoNothing && tx.RowsAffected == 0:
		default:
			if err := opts.withAfterWrite(i); err != nil {
				return fmt.Errorf("%s: error after write: %w", op, err)
			}
		}
	}
	if err := rw.lookupAfterWrite(ctx, i, opt...); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// CreateItems will create multiple items of the same type. Supported options:
// WithDebug, WithBeforeWrite, WithAfterWrite, WithReturnRowsAffected,
// OnConflict, WithVersion, and WithWhere. WithLookup is not a supported option.
func (rw *RW) CreateItems(ctx context.Context, createItems []interface{}, opt ...Option) error {
	const op = "db.CreateItems"
	if rw.underlying == nil {
		return fmt.Errorf("%s: missing underlying db: %w", op, ErrInvalidParameter)
	}
	if len(createItems) == 0 {
		return fmt.Errorf("%s: missing interfaces: %w", op, ErrInvalidParameter)
	}
	opts := getOpts(opt...)
	if opts.withLookup {
		return fmt.Errorf("%s: with lookup not a supported option: %w", op, ErrInvalidParameter)
	}
	// verify that createItems are all the same type.
	var foundType reflect.Type
	for i, v := range createItems {
		if i == 0 {
			foundType = reflect.TypeOf(v)
		}
		currentType := reflect.TypeOf(v)
		if foundType != currentType {
			return fmt.Errorf("%s: create items contains disparate types. item %d is not a %s: %w", op, i, foundType.Name(), ErrInvalidParameter)
		}
	}
	if opts.withBeforeWrite != nil {
		if err := opts.withBeforeWrite(createItems); err != nil {
			return fmt.Errorf("%s: error before write: %w", op, err)
		}
	}
	for _, item := range createItems {
		if err := rw.Create(ctx, item,
			WithOnConflict(opts.withOnConflict),
			WithReturnRowsAffected(opts.withRowsAffected),
			WithDebug(opts.withDebug),
			WithVersion(opts.WithVersion),
			WithWhere(opts.withWhereClause, opts.withWhereClauseArgs...),
		); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	if opts.withAfterWrite != nil {
		if err := opts.withAfterWrite(createItems); err != nil {
			return fmt.Errorf("%s: error after write: %w", op, err)
		}
	}
	return nil
}
