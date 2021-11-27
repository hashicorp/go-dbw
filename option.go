package dbw

import (
	"github.com/hashicorp/go-hclog"
)

// GetOpts - iterate the inbound Options and return a struct.
func GetOpts(opt ...Option) Options {
	opts := getDefaultOptions()
	for _, o := range opt {
		if o != nil {
			o(&opts)
		}
	}
	return opts
}

// Option - how Options are passed as arguments.
type Option func(*Options)

// Options - how Options are represented.
type Options struct {
	withBeforeWrite func(i interface{}) error
	withAfterWrite  func(i interface{}, rowsAffected int) error

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
	withMinOpenConnections int

	// withDebug indicates that the given operation should invoke debug output
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
// write operation. The i interface{} passed at runtime will be the resource(s)
// being written.
func WithBeforeWrite(fn func(i interface{}) error) Option {
	return func(o *Options) {
		o.withBeforeWrite = fn
	}
}

// WithAfterWrite provides and option to provide a func to be called after a
// write operation.  The i interface{} passed at runtime will be the resource(s)
// being written.
func WithAfterWrite(fn func(i interface{}, rowsAffected int) error) Option {
	return func(o *Options) {
		o.withAfterWrite = fn
	}
}

// WithLookup enables a lookup after a write operation.
func WithLookup(enable bool) Option {
	return func(o *Options) {
		o.withLookup = enable
	}
}

// WithFieldMaskPaths provides an option to provide field mask paths for update
// operations.
func WithFieldMaskPaths(paths []string) Option {
	return func(o *Options) {
		o.WithFieldMaskPaths = paths
	}
}

// WithNullPaths provides an option to provide null paths for update operations.
//
func WithNullPaths(paths []string) Option {
	return func(o *Options) {
		o.WithNullPaths = paths
	}
}

// WithLimit provides an option to provide a limit.  Intentionally allowing
// negative integers.   If WithLimit < 0, then unlimited results are returned.
// If WithLimit == 0, then default limits are used for results (see DefaultLimit
// const).
func WithLimit(limit int) Option {
	return func(o *Options) {
		o.WithLimit = limit
	}
}

// WithVersion provides an option version number for update operations.  Using
// this option requires that your resource has a version column that's
// incremented for every successful update operation.  Version provides an
// optimistic locking mechanism for write operations.
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

// WithLogger specifies an optional hclog to use for db operations.  It's only
// valid for Open(..) and OpenWith(...)
func WithLogger(l hclog.Logger) Option {
	return func(o *Options) {
		o.withGormFormatter = l
	}
}

// WithMaxOpenConnections specifices and optional max open connections for the
// database.  A value of zero equals unlimited connections
func WithMaxOpenConnections(max int) Option {
	return func(o *Options) {
		o.withMaxOpenConnections = max
	}
}

// WithMinOpenConnections specifices and optional min open connections for the
// database.  A value of zero means that there is no min.
func WithMinOpenConnections(max int) Option {
	return func(o *Options) {
		o.withMinOpenConnections = max
	}
}

// WithDebug specifies the given operation should invoke debug mode for the
// database output
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
// and typically used with "bulk" write operations.
func WithReturnRowsAffected(rowsAffected *int64) Option {
	return func(o *Options) {
		o.withRowsAffected = rowsAffected
	}
}
