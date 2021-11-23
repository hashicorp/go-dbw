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
	gormDb := rw.underlying.Exec(sql, values...)
	if gormDb.Error != nil {
		return NoRowsAffected, fmt.Errorf("%s: %w", op, gormDb.Error)
	}
	return int(gormDb.RowsAffected), nil
}

func setFieldsToNil(i interface{}, fieldNames []string) {
	// Note: error cases are not handled
	_ = Clear(i, fieldNames, 2)
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

// Clear sets fields in the value pointed to by i to their zero value.
// Clear descends i to depth clearing fields at each level. i must be a
// pointer to a struct. Cycles in i are not detected.
//
// A depth of 2 will change i and i's children. A depth of 1 will change i
// but no children of i. A depth of 0 will return with no changes to i.
func Clear(i interface{}, fields []string, depth int) error {
	const op = "dbw.Clear"
	if len(fields) == 0 || depth == 0 {
		return nil
	}
	fm := make(map[string]bool)
	for _, f := range fields {
		fm[f] = true
	}

	v := reflect.ValueOf(i)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() || v.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("%s: %w", op, ErrInvalidParameter)
		}
		clear(v, fm, depth)
	default:
		return fmt.Errorf("%s: %w", op, ErrInvalidParameter)
	}
	return nil
}

func clear(v reflect.Value, fields map[string]bool, depth int) {
	if depth == 0 {
		return
	}
	depth--

	switch v.Kind() {
	case reflect.Ptr:
		clear(v.Elem(), fields, depth+1)
	case reflect.Struct:
		typeOfT := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if ok := fields[typeOfT.Field(i).Name]; ok {
				if f.IsValid() && f.CanSet() {
					f.Set(reflect.Zero(f.Type()))
				}
				continue
			}
			clear(f, fields, depth)
		}
	}
}

func (rw *RW) whereClausesFromOpts(ctx context.Context, i interface{}, opts Options) (string, []interface{}, error) {
	const op = "dbw.whereClausesFromOpts"
	var where []string
	var args []interface{}
	if opts.WithVersion != nil {
		if *opts.WithVersion == 0 {
			return "", nil, fmt.Errorf("%s: with version option is zero: %w", op, ErrInvalidParameter)
		}
		mDb := rw.underlying.Model(i)
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
	tx := rw.underlying.Model(i)
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

func (_ *RW) Update(ctx context.Context, i interface{}, fieldMaskPaths []string, setToNullPaths []string, opt ...Option) (int, error) {
	panic("todo")
}

func (_ *RW) CreateItems(ctx context.Context, createItems []interface{}, opt ...Option) error {
	panic("todo")
}

func (_ *RW) Delete(ctx context.Context, i interface{}, opt ...Option) (int, error) {
	panic("todo")
}

func (_ *RW) DeleteItems(ctx context.Context, deleteItems []interface{}, opt ...Option) (int, error) {
	panic("todo")
}
