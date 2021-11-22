package dbw

import (
	"github.com/hashicorp/go-hclog"
)

// getOpts - iterate the inbound Options and return a struct.
func getOpts(opt ...Option) Options {
	opts := getDefaultOptions()
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

// Option - how Options are passed as arguments.
type Option func(*Options)

// Options - how Options are represented.
type Options struct {
	withBeforeWrite func() error
	withAfterWrite  func() error

	withLookup bool
	// WithLimit must be accessible in other packages.
	WithLimit int
	// WithFieldMaskPaths must be accessible from other packages.
	WithFieldMaskPaths []string
	// WithNullPaths must be accessible from other packages.
	WithNullPaths []string

	// WithVersion must be accessible from other packages.
	WithVersion *uint32

	withSkipVetForWrite bool

	withWhereClause     string
	withWhereClauseArgs []interface{}
	withOrder           string

	// withPrngValues is used to switch the ID generation to a pseudo-random mode
	withPrngValues []string

	withGormFormatter      hclog.Logger
	withMaxOpenConnections int

	// withDebug indicates that the given operation should invoke Gorm's debug
	// mode
	withDebug bool

	withOnConflict   *OnConflict
	withRowsAffected *int64
}

func getDefaultOptions() Options {
	return Options{
		WithFieldMaskPaths: []string{},
		WithNullPaths:      []string{},
	}
}

// WithBeforeWrite provides and option to provide a func to be called before a
// write operation.
func WithBeforeWrite(fn func() error) Option {
	return func(o *Options) {
		o.withBeforeWrite = fn
	}
}

// WithAfterWrite provides and option to provide a func to be called after a
// write operation
func WithAfterWrite(fn func() error) Option {
	return func(o *Options) {
		o.withAfterWrite = fn
	}
}

// WithLookup enables a lookup.
func WithLookup(enable bool) Option {
	return func(o *Options) {
		o.withLookup = enable
	}
}

// WithFieldMaskPaths provides an option to provide field mask paths.
func WithFieldMaskPaths(paths []string) Option {
	return func(o *Options) {
		o.WithFieldMaskPaths = paths
	}
}

// WithNullPaths provides an option to provide null paths.
func WithNullPaths(paths []string) Option {
	return func(o *Options) {
		o.WithNullPaths = paths
	}
}

// WithLimit provides an option to provide a limit.  Intentionally allowing
// negative integers.   If WithLimit < 0, then unlimited results are returned.
// If WithLimit == 0, then default limits are used for results.
func WithLimit(limit int) Option {
	return func(o *Options) {
		o.WithLimit = limit
	}
}

// WithVersion provides an option version number for update operations.
func WithVersion(version *uint32) Option {
	return func(o *Options) {
		o.WithVersion = version
	}
}

// WithSkipVetForWrite provides an option to allow skipping vet checks to allow
// testing lower-level SQL triggers and constraints
func WithSkipVetForWrite(enable bool) Option {
	return func(o *Options) {
		o.withSkipVetForWrite = enable
	}
}

// WithWhere provides an option to provide a where clause with arguments for an
// operation.
func WithWhere(whereClause string, args ...interface{}) Option {
	return func(o *Options) {
		o.withWhereClause = whereClause
		o.withWhereClauseArgs = append(o.withWhereClauseArgs, args...)
	}
}

// WithOrder provides an option to provide an order when searching and looking
// up.
func WithOrder(withOrder string) Option {
	return func(o *Options) {
		o.withOrder = withOrder
	}
}

// WithPrngValues provides an option to provide values to seed an PRNG when generating IDs
func WithPrngValues(withPrngValues []string) Option {
	return func(o *Options) {
		o.withPrngValues = withPrngValues
	}
}

// WithGormFormatter specifies an optional hclog to use for gorm's log
// formmater
func WithGormFormatter(l hclog.Logger) Option {
	return func(o *Options) {
		o.withGormFormatter = l
	}
}

// WithMaxOpenConnections specifices and optional max open connections for the
// database
func WithMaxOpenConnections(max int) Option {
	return func(o *Options) {
		o.withMaxOpenConnections = max
	}
}

// WithDebug specifies the given operation should invoke debug mode in Gorm
func WithDebug(with bool) Option {
	return func(o *Options) {
		o.withDebug = with
	}
}

// WithOnConflict specifies an optional on conflict criteria which specify
// alternative actions to take when an insert results in a unique constraint or
// exclusion constraint error
func WithOnConflict(onConflict *OnConflict) Option {
	return func(o *Options) {
		o.withOnConflict = onConflict
	}
}

// WithReturnRowsAffected specifies an option for returning the rows affected
func WithReturnRowsAffected(rowsAffected *int64) Option {
	return func(o *Options) {
		o.withRowsAffected = rowsAffected
	}
}
