package dbw

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

const (
	NoRowsAffected = 0

	// DefaultLimit is the default for results for boundary
	DefaultLimit = 10000
)

// RW uses a DB connection for read/write
type RW struct {
	underlying *DB
}

// ensure that RW implements the interfaces of: Reader and Writer
var (
	_ Reader = (*RW)(nil)
	_ Writer = (*RW)(nil)
)

func New(underlying *DB) *RW {
	return &RW{underlying: underlying}
}

// Exec will execute the sql with the values as parameters. The int returned
// is the number of rows affected by the sql. No options are currently
// supported.
func (rw *RW) Exec(ctx context.Context, sql string, values []interface{}, _ ...Option) (int, error) {
	const op = "dbw.Exec"
	if rw.underlying == nil {
		return 0, fmt.Errorf("%s: missing underlying db: %w", op, ErrInternal)
	}
	if sql == "" {
		return NoRowsAffected, fmt.Errorf("%s: missing sql: %w", op, ErrInvalidParameter)
	}
	db := rw.underlying.wrapped.Exec(sql, values...)
	if db.Error != nil {
		return NoRowsAffected, fmt.Errorf("%s: %w", op, db.Error)
	}
	return int(db.RowsAffected), nil
}

func (rw *RW) primaryFieldsAreZero(ctx context.Context, i interface{}) ([]string, bool, error) {
	const op = "db.primaryFieldsAreZero"
	var fieldNames []string
	tx := rw.underlying.wrapped.Model(i)
	if err := tx.Statement.Parse(i); err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, ErrInvalidParameter)
	}
	for _, f := range tx.Statement.Schema.PrimaryFields {
		if f.PrimaryKey {
			if _, isZero := f.ValueOf(reflect.ValueOf(i)); isZero {
				fieldNames = append(fieldNames, f.Name)
			}
		}
	}
	return fieldNames, len(fieldNames) > 0, nil
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func contains(ss []string, t string) bool {
	for _, s := range ss {
		if strings.EqualFold(s, t) {
			return true
		}
	}
	return false
}

func (rw *RW) whereClausesFromOpts(ctx context.Context, i interface{}, opts Options) (string, []interface{}, error) {
	const op = "dbw.whereClausesFromOpts"
	var where []string
	var args []interface{}
	if opts.WithVersion != nil {
		if *opts.WithVersion == 0 {
			return "", nil, fmt.Errorf("%s: with version option is zero: %w", op, ErrInvalidParameter)
		}
		mDb := rw.underlying.wrapped.Model(i)
		err := mDb.Statement.Parse(i)
		if err != nil && mDb.Statement.Schema == nil {
			return "", nil, fmt.Errorf("%s: (internal error) unable to parse stmt: %w", op, ErrUnknown)
		}
		if !contains(mDb.Statement.Schema.DBNames, "version") {
			return "", nil, fmt.Errorf("%s: %s does not have a version field: %w", op, mDb.Statement.Schema.Table, ErrInvalidParameter)
		}
		where = append(where, fmt.Sprintf("%s.version = ?", mDb.Statement.Schema.Table)) // we need to include the table name because of "on conflict" use cases
		args = append(args, opts.WithVersion)
	}
	if opts.withWhereClause != "" {
		where, args = append(where, opts.withWhereClause), append(args, opts.withWhereClauseArgs...)
	}
	return strings.Join(where, " and "), args, nil
}

func (rw *RW) primaryKeysWhere(ctx context.Context, i interface{}) (string, []interface{}, error) {
	const op = "dbw.primaryKeysWhere"
	var fieldNames []string
	var fieldValues []interface{}
	tx := rw.underlying.wrapped.Model(i)
	if err := tx.Statement.Parse(i); err != nil {
		return "", nil, fmt.Errorf("%s: %w", op, err)
	}
	switch resourceType := i.(type) {
	case ResourcePublicIder:
		if resourceType.GetPublicId() == "" {
			return "", nil, fmt.Errorf("%s: missing primary key: %w", op, ErrInvalidParameter)
		}
		fieldValues = []interface{}{resourceType.GetPublicId()}
		fieldNames = []string{"public_id"}
	case ResourcePrivateIder:
		if resourceType.GetPrivateId() == "" {
			return "", nil, fmt.Errorf("%s: missing primary key: %w", op, ErrInvalidParameter)
		}
		fieldValues = []interface{}{resourceType.GetPrivateId()}
		fieldNames = []string{"private_id"}
	default:
		v := reflect.ValueOf(i)
		for _, f := range tx.Statement.Schema.PrimaryFields {
			if f.PrimaryKey {
				val, isZero := f.ValueOf(v)
				if isZero {
					return "", nil, fmt.Errorf("%s: primary field %s is zero: %w", op, f.Name, ErrInvalidParameter)
				}
				fieldNames = append(fieldNames, f.DBName)
				fieldValues = append(fieldValues, val)
			}
		}
	}
	if len(fieldNames) == 0 {
		return "", nil, fmt.Errorf("%s: no primary key(s) for %t: %w", op, i, ErrInvalidParameter)
	}
	clauses := make([]string, 0, len(fieldNames))
	for _, col := range fieldNames {
		clauses = append(clauses, fmt.Sprintf("%s = ?", col))
	}
	return strings.Join(clauses, " and "), fieldValues, nil
}

func (_ *RW) LookupWhere(ctx context.Context, resource interface{}, where string, args ...interface{}) error {
	panic("todo")
}

func (_ *RW) SearchWhere(ctx context.Context, resources interface{}, where string, args []interface{}, opt ...Option) error {
	panic("todo")
}

func (_ *RW) Delete(ctx context.Context, i interface{}, opt ...Option) (int, error) {
	panic("todo")
}

func (_ *RW) DeleteItems(ctx context.Context, deleteItems []interface{}, opt ...Option) (int, error) {
	panic("todo")
}
