package dbw_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDb_Create(t *testing.T) {
	testCtx := context.Background()
	db, _ := dbw.TestSetup(t)
	t.Run("simple", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		id, err := dbw.NewId("u")
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		ts := &dbtest.Timestamp{Timestamp: timestamppb.Now()}
		user.CreateTime = ts
		user.UpdateTime = ts
		user.Name = "alice-" + id
		err = w.Create(testCtx, user, dbw.WithDebug(true))
		require.NoError(err)
		assert.NotEmpty(user.PublicId)
		// make sure the database controlled the timestamp values
		assert.NotEqual(ts, user.GetCreateTime())
		assert.NotEqual(ts, user.GetUpdateTime())

		foundUser, err := dbtest.NewTestUser()
		require.NoError(err)
		foundUser.PublicId = user.PublicId
		err = w.LookupByPublicId(testCtx, foundUser)
		require.NoError(err)
		assert.Equal(foundUser.PublicId, user.PublicId)
	})
	t.Run("WithBeforeCreate", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		id, err := dbw.NewId("u")
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "alice" + id
		fn := func(i interface{}) error {
			u, ok := i.(*dbtest.TestUser)
			require.True(ok)
			u.Name = "before" + id
			return nil
		}
		err = w.Create(
			testCtx,
			user,
			dbw.WithBeforeWrite(fn),
		)
		require.NoError(err)
		require.NotEmpty(user.PublicId)
		require.Equal("before"+id, user.Name)

		foundUser, err := dbtest.NewTestUser()
		require.NoError(err)
		foundUser.PublicId = user.PublicId
		err = w.LookupByPublicId(testCtx, foundUser)
		require.NoError(err)
		assert.Equal(foundUser.PublicId, user.PublicId)
		assert.Equal("before"+id, foundUser.Name)

		fn = func(i interface{}) error {
			return errors.New("fail")
		}
		err = w.Create(
			testCtx,
			user,
			dbw.WithBeforeWrite(fn),
		)
		require.Error(err)

	})
	t.Run("WithAfterCreate", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		db.Debug(true)
		id, err := dbw.NewId("u")
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "alice" + id
		fn := func(i interface{}, rowAffected int) error {
			u, ok := i.(*dbtest.TestUser)
			require.True(ok)
			rowsAffected, err := w.Exec(testCtx,
				"update db_test_user set name = @name where public_id = @public_id",
				[]interface{}{
					sql.Named("name", "after"+id),
					sql.Named("public_id", u.PublicId),
				})
			require.NoError(err)
			require.Equal(1, rowsAffected)
			// since we're going to use WithLookup(true), we don't need to set
			// name here.
			return nil
		}
		err = w.Create(
			context.Background(),
			user,
			dbw.WithAfterWrite(fn),
			dbw.WithLookup(true),
		)
		require.NoError(err)
		require.NotEmpty(user.PublicId)
		require.Equal("after"+id, user.Name)

		foundUser, err := dbtest.NewTestUser()
		require.NoError(err)
		foundUser.PublicId = user.PublicId
		err = w.LookupByPublicId(context.Background(), foundUser)
		require.NoError(err)
		assert.Equal(foundUser.PublicId, user.PublicId)
		assert.Equal("after"+id, foundUser.Name)

		fn = func(i interface{}, rowsAffected int) error {
			return errors.New("fail")
		}

		user2, err := dbtest.NewTestUser()
		require.NoError(err)
		err = w.Create(
			context.Background(),
			user2,
			dbw.WithAfterWrite(fn),
		)
		require.Error(err)

	})
	t.Run("nil-tx", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(nil)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		err = w.Create(context.Background(), user)
		require.Error(err)
		assert.Contains(err.Error(), "dbw.Create: missing underlying db: invalid parameter")
	})
	t.Run("nil-resource", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		err := w.Create(context.Background(), nil)
		require.Error(err)
		assert.Contains(err.Error(), "dbw.Create: missing interface: invalid parameter")
	})
	t.Run("VetForWrite-err", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		u := dbtest.AllocTestUser()
		err := w.Create(context.Background(), &u)
		require.Error(err)
		assert.Contains(err.Error(), "dbtest.(TestUser).VetForWrite: missing public id: invalid parameter")
	})
}

func TestDb_Create_OnConflict(t *testing.T) {
	ctx := context.Background()
	conn, _ := dbw.TestSetup(t)
	rw := dbw.New(conn)
	dbType, err := conn.DbType()
	require.NoError(t, err)

	createInitialUser := func() *dbtest.TestUser {
		// create initial user for on conflict tests
		id, err := dbw.NewId("test-user")
		require.NoError(t, err)
		initialUser, err := dbtest.NewTestUser()
		require.NoError(t, err)
		ts := &dbtest.Timestamp{Timestamp: timestamppb.Now()}
		initialUser.CreateTime = ts
		initialUser.UpdateTime = ts
		initialUser.Name = "foo-" + id
		err = rw.Create(ctx, initialUser)
		require.NoError(t, err)
		assert.NotEmpty(t, initialUser.PublicId)
		assert.Equal(t, uint32(1), initialUser.Version)
		return initialUser
	}

	tests := []struct {
		name            string
		onConflict      dbw.OnConflict
		additionalOpts  []dbw.Option
		wantUpdate      bool
		wantEmail       string
		withDebug       bool
		wantErrContains string
	}{
		{
			name: "invalid-target",
			onConflict: dbw.OnConflict{
				Target: "invalid",
				Action: dbw.SetColumns([]string{"name"}),
			},
			wantErrContains: "dbw.Create: invalid conflict target string: invalid parameter",
		},
		{
			name: "invalid-action",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: "invalid",
			},
			wantErrContains: "dbw.Create: invalid conflict action string: invalid parameter",
		},
		{
			name: "set-columns",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			wantUpdate: true,
		},
		{
			name: "set-column-values",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumnValues(map[string]interface{}{
					"name":         dbw.Expr("lower(?)", "alice eve smith"),
					"email":        "alice@gmail.com",
					"phone_number": dbw.Expr("NULL"),
				}),
			},
			wantUpdate: true,
			wantEmail:  "alice@gmail.com",
		},
		{
			name: "both-set-columns-and-set-column-values",
			onConflict: func() dbw.OnConflict {
				onConflict := dbw.OnConflict{
					Target: dbw.Columns{"public_id"},
				}
				cv := dbw.SetColumns([]string{"name"})
				cv = append(cv,
					dbw.SetColumnValues(map[string]interface{}{
						"email":        "alice@gmail.com",
						"phone_number": dbw.Expr("NULL"),
					})...)
				onConflict.Action = cv
				return onConflict
			}(),
			wantUpdate: true,
			wantEmail:  "alice@gmail.com",
		},
		{
			name: "do-nothing",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.DoNothing(true),
			},
			wantUpdate: false,
		},
		{
			name: "on-constraint",
			onConflict: dbw.OnConflict{
				Target: dbw.Constraint("db_test_user_pkey"),
				Action: dbw.SetColumns([]string{"name"}),
			},
			wantUpdate: true,
		},
		{
			name: "set-columns-with-where-success",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			additionalOpts: []dbw.Option{dbw.WithWhere("db_test_user.version = ?", 1)},
			wantUpdate:     true,
		},
		{
			name: "set-columns-with-where-fail",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			additionalOpts: []dbw.Option{dbw.WithWhere("db_test_user.version = ?", 100000000000)},
			wantUpdate:     false,
		},
		{
			name: "set-columns-with-version-success",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			additionalOpts: []dbw.Option{dbw.WithVersion(func() *uint32 { i := uint32(1); return &i }())},
			wantUpdate:     true,
		},
		{
			name: "set-columns-with-version-fail",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			additionalOpts: []dbw.Option{dbw.WithWhere("db_test_user.version = ?", 100000000000)},
			wantUpdate:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if dbType == dbw.Sqlite {
				// sqlite doesn't support "on conflict on constraint" targets
				if _, ok := tt.onConflict.Target.(dbw.Constraint); ok {
					return
				}
			}
			assert, require := assert.New(t), require.New(t)
			initialUser := createInitialUser()
			conflictUser, err := dbtest.NewTestUser()
			require.NoError(err)
			userNameId, err := dbw.NewId("test-user-name")
			require.NoError(err)
			conflictUser.PublicId = initialUser.PublicId
			conflictUser.Name = userNameId
			var rowsAffected int64
			opts := []dbw.Option{dbw.WithOnConflict(&tt.onConflict), dbw.WithReturnRowsAffected(&rowsAffected)}
			if tt.additionalOpts != nil {
				opts = append(opts, tt.additionalOpts...)
			}
			if tt.withDebug {
				conn.Debug(true)
			}
			err = rw.Create(ctx, conflictUser, opts...)
			if tt.withDebug {
				conn.Debug(false)
			}
			if tt.wantErrContains != "" {
				require.Error(err)
				assert.Contains(err.Error(), tt.wantErrContains)
				return
			}
			require.NoError(err)
			foundUser, err := dbtest.NewTestUser()
			require.NoError(err)
			foundUser.PublicId = conflictUser.PublicId
			err = rw.LookupByPublicId(context.Background(), foundUser)
			require.NoError(err)
			t.Log(foundUser)
			if tt.wantUpdate {
				assert.Equal(int64(1), rowsAffected)
				assert.Equal(conflictUser.PublicId, foundUser.PublicId)
				assert.Equal(conflictUser.Name, foundUser.Name)
				if tt.wantEmail != "" {
					assert.Equal(tt.wantEmail, foundUser.Email)
				}
			} else {
				assert.Equal(int64(0), rowsAffected)
				assert.Equal(conflictUser.PublicId, foundUser.PublicId)
				assert.NotEqual(conflictUser.Name, foundUser.Name)
			}
		})
	}
	t.Run("update-all", func(t *testing.T) {
		// for now, let's just deal with postgres, since all dialects are a
		// bit diff when it comes to auto-incremented pks.  Also, gorm currently
		// is great in "RETURNING WITH" for auto-incremented keys for sqlite
		if dbType != dbw.Postgres {
			return
		}

		assert, require := assert.New(t), require.New(t)
		// we need a table with an auto-increment pk for update all
		const createTable = `create table if not exists db_test_update_alls (
			id bigint generated always as identity primary key,
			public_id text not null unique,
			name text unique,
			phone_number text,
			email text
		  )`

		_, err := rw.Exec(context.Background(), createTable, nil)
		require.NoError(err)

		// create initial resource for the test
		id, err := dbw.NewId("test")
		require.NoError(err)
		initialResource := &dbTestUpdateAll{
			PublicId: id,
			Name:     "foo-" + id,
		}
		conn.Debug(true)
		err = rw.Create(ctx, initialResource)
		require.NoError(err)
		assert.NotEmpty(initialResource.PublicId)

		nameId, err := dbw.NewId("test-name")
		require.NoError(err)
		conflictResource := &dbTestUpdateAll{
			PublicId: id,
			Name:     nameId,
		}
		onConflict := dbw.OnConflict{
			Target: dbw.Columns{"public_id"},
			Action: dbw.UpdateAll(true),
		}
		var rowsAffected int64
		opts := []dbw.Option{dbw.WithOnConflict(&onConflict), dbw.WithReturnRowsAffected(&rowsAffected)}
		err = rw.Create(ctx, conflictResource, opts...)

		require.NoError(err)
		foundResource := &dbTestUpdateAll{
			PublicId: conflictResource.PublicId,
		}
		rw.LookupByPublicId(context.Background(), foundResource)
		t.Log(foundResource)
		require.NoError(err)
		assert.Equal(int64(1), rowsAffected)
		assert.Equal(conflictResource.PublicId, foundResource.PublicId)
		assert.Equal(conflictResource.Name, foundResource.Name)
	})
	t.Run("CreateItems", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		initialUser := createInitialUser()
		conflictUser, err := dbtest.NewTestUser()
		require.NoError(err)
		userNameId, err := dbw.NewId("test-user-name")
		require.NoError(err)
		conflictUser.PublicId = initialUser.PublicId
		conflictUser.Name = userNameId
		onConflict := dbw.OnConflict{
			Target: dbw.Columns{"public_id"},
			Action: dbw.SetColumns([]string{"name"}),
		}
		users := []interface{}{}
		users = append(users, conflictUser)
		var rowsAffected int64
		err = rw.CreateItems(ctx, users, dbw.WithOnConflict(&onConflict), dbw.WithReturnRowsAffected(&rowsAffected))
		require.NoError(err)
		foundUser, err := dbtest.NewTestUser()
		require.NoError(err)
		foundUser.PublicId = conflictUser.PublicId
		err = rw.LookupByPublicId(context.Background(), foundUser)
		require.NoError(err)

		assert.Equal(int64(1), rowsAffected)
		assert.Equal(conflictUser.PublicId, foundUser.PublicId)
		assert.Equal(conflictUser.Name, foundUser.Name)
	})
}

func TestDb_CreateItems(t *testing.T) {
	testCtx := context.Background()
	conn, _ := dbw.TestSetup(t)
	testRw := dbw.New(conn)

	createFn := func() []interface{} {
		results := []interface{}{}
		for i := 0; i < 10; i++ {
			u, err := dbtest.NewTestUser()
			require.NoError(t, err)
			results = append(results, u)
		}
		return results
	}
	createMixedFn := func() []interface{} {
		u, err := dbtest.NewTestUser()
		require.NoError(t, err)
		c, err := dbtest.NewTestCar()
		require.NoError(t, err)
		return []interface{}{
			u,
			c,
		}
	}
	successBeforeFn := func(_ interface{}) error {
		return nil
	}
	successAfterFn := func(_ interface{}, _ int) error {
		return nil
	}
	errFailedFn := errors.New("fail")
	failedBeforeFn := func(_ interface{}) error {
		return errFailedFn
	}
	failedAfterFn := func(_ interface{}, _ int) error {
		return errFailedFn
	}
	type args struct {
		createItems []interface{}
		opt         []dbw.Option
	}
	tests := []struct {
		name      string
		rw        *dbw.RW
		args      args
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "simple",
			rw:   testRw,
			args: args{
				createItems: createFn(),
			},
			wantErr: false,
		},
		{
			name: "simple-with-before-after-success",
			rw:   testRw,
			args: args{
				createItems: createFn(),
				opt: []dbw.Option{
					dbw.WithBeforeWrite(successBeforeFn),
					dbw.WithAfterWrite(successAfterFn),
				},
			},
			wantErr: false,
		},
		{
			name: "failed-with-before",
			rw:   testRw,
			args: args{
				createItems: createFn(),
				opt: []dbw.Option{
					dbw.WithBeforeWrite(failedBeforeFn),
				},
			},
			wantErr:   true,
			wantErrIs: errFailedFn,
		},
		{
			name: "failed-with-after",
			rw:   testRw,
			args: args{
				createItems: createFn(),
				opt: []dbw.Option{
					dbw.WithAfterWrite(failedAfterFn),
				},
			},
			wantErr:   true,
			wantErrIs: errFailedFn,
		},
		{
			name: "mixed items",
			rw:   testRw,
			args: args{
				createItems: createMixedFn(),
			},
			wantErr:   true,
			wantErrIs: dbw.ErrInvalidParameter,
		},
		{
			name: "bad opt: WithLookup",
			rw:   testRw,
			args: args{
				createItems: createFn(),
				opt:         []dbw.Option{dbw.WithLookup(true)},
			},
			wantErr:   true,
			wantErrIs: dbw.ErrInvalidParameter,
		},
		{
			name: "nil underlying",
			rw:   &dbw.RW{},
			args: args{
				createItems: createFn(),
			},
			wantErr:   true,
			wantErrIs: dbw.ErrInvalidParameter,
		},
		{
			name: "empty items",
			rw:   testRw,
			args: args{
				createItems: []interface{}{},
			},
			wantErr:   true,
			wantErrIs: dbw.ErrInvalidParameter,
		},
		{
			name: "nil items",
			rw:   testRw,
			args: args{
				createItems: nil,
			},
			wantErr:   true,
			wantErrIs: dbw.ErrInvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			err := tt.rw.CreateItems(testCtx, tt.args.createItems, tt.args.opt...)
			if tt.wantErr {
				require.Error(err)
				assert.ErrorIsf(err, tt.wantErrIs, "unexpected error: %s", err.Error())
				return
			}
			require.NoError(err)
			for _, item := range tt.args.createItems {
				u := dbtest.AllocTestUser()
				u.PublicId = item.(*dbtest.TestUser).PublicId
				err := tt.rw.LookupByPublicId(context.Background(), &u)
				assert.NoError(err)
				if _, ok := item.(*dbtest.TestUser); ok {
					assert.Truef(proto.Equal(item.(*dbtest.TestUser).StoreTestUser, u.StoreTestUser), "%s and %s should be equal", item, u)
				}
			}
		})
	}
}

type dbTestUpdateAll struct {
	Id          int `gorm:"primary_key"`
	PublicId    string
	Name        string `gorm:"default:null"`
	PhoneNumber string `gorm:"default:null"`
	Email       string `gorm:"default:null"`
}

func (r *dbTestUpdateAll) GetPublicId() string {
	return r.PublicId
}
