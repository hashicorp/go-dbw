// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dbw_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strconv"
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
		err = w.Create(testCtx, user)
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
	t.Run("WithBeforeWrite", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		id, err := dbw.NewId("u")
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "alice" + id
		fn := func(i any) error {
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

		fn = func(i any) error {
			return errors.New("fail")
		}
		err = w.Create(
			testCtx,
			user,
			dbw.WithBeforeWrite(fn),
		)
		require.Error(err)
	})
	t.Run("WithAfterWrite", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		id, err := dbw.NewId("u")
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "alice" + id
		fn := func(i any, rowAffected int) error {
			u, ok := i.(*dbtest.TestUser)
			require.True(ok)
			rowsAffected, err := w.Exec(testCtx,
				"update db_test_user set name = @name where public_id = @public_id",
				[]any{
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

		fn = func(i any, rowsAffected int) error {
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
	t.Run("hooks", func(t *testing.T) {
		hookTests := []struct {
			name     string
			resource any
		}{
			{"before-create", &dbtest.TestWithBeforeCreate{}},
			{"after-create", &dbtest.TestWithAfterCreate{}},
			{"before-save", &dbtest.TestWithBeforeSave{}},
			{"before-save", &dbtest.TestWithAfterSave{}},
		}
		for _, tt := range hookTests {
			t.Run(tt.name, func(t *testing.T) {
				assert, require := assert.New(t), require.New(t)
				w := dbw.New(db)
				err := w.Create(context.Background(), tt.resource)
				require.Error(err)
				assert.ErrorIs(err, dbw.ErrInvalidParameter)
				assert.Contains(err.Error(), "gorm callback/hooks are not supported")
			})
		}
	})
	t.Run("WithTable", func(t *testing.T) {
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
		err = w.Create(testCtx, user, dbw.WithTable(user.TableName()))
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

		user2, err := dbtest.NewTestUser()
		require.NoError(err)
		err = w.Create(testCtx, user2, dbw.WithTable("invalid-table"))
		require.Error(err)

		err = w.Create(testCtx, user2, dbw.WithTable(user.TableName()))
		require.NoError(err)
	})
}

func TestDb_Create_OnConflict(t *testing.T) {
	ctx := context.Background()
	conn, _ := dbw.TestSetup(t)
	rw := dbw.New(conn)
	dbType, _, err := conn.DbType()
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
				Action: dbw.SetColumnValues(map[string]any{
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
					dbw.SetColumnValues(map[string]any{
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
}

func TestDb_CreateItems(t *testing.T) {
	testCtx := context.Background()
	conn, _ := dbw.TestSetup(t)
	testRw := dbw.New(conn)

	testWithTableUser, err := dbtest.NewTestUser()
	require.NoError(t, err)

	createFn := func() any {
		results := []*dbtest.TestUser{}
		for i := 0; i < 10; i++ {
			u, err := dbtest.NewTestUser()
			require.NoError(t, err)
			results = append(results, u)
		}
		return results
	}
	createMixedFn := func() []any {
		u, err := dbtest.NewTestUser()
		require.NoError(t, err)
		c, err := dbtest.NewTestCar()
		require.NoError(t, err)
		return []any{
			u,
			c,
		}
	}
	successBeforeFn := func(_ any) error {
		return nil
	}
	successAfterFn := func(_ any, _ int) error {
		return nil
	}
	errFailedFn := errors.New("fail")
	failedBeforeFn := func(_ any) error {
		return errFailedFn
	}
	failedAfterFn := func(_ any, _ int) error {
		return errFailedFn
	}
	type args struct {
		createItems any
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
			name: "with-table",
			rw:   testRw,
			args: args{
				createItems: createFn(),
				opt:         []dbw.Option{dbw.WithTable(testWithTableUser.TableName())},
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
				createItems: []any{},
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
		{
			name: "not a slice",
			rw:   testRw,
			args: args{
				createItems: "not a slice",
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
			val := reflect.ValueOf(tt.args.createItems)
			for i := 0; i < val.Len(); i++ {
				u := dbtest.AllocTestUser()
				u.PublicId = val.Index(i).Interface().(*dbtest.TestUser).PublicId
				err := tt.rw.LookupByPublicId(context.Background(), &u)
				assert.NoError(err)
				if _, ok := val.Index(i).Interface().(*dbtest.TestUser); ok {
					assert.Truef(proto.Equal(val.Index(i).Interface().(*dbtest.TestUser).StoreTestUser, u.StoreTestUser), "%s and %s should be equal", val.Index(i).Interface().(*dbtest.TestUser), u)
				}
			}
		})
	}
	t.Run("hooks", func(t *testing.T) {
		hookTests := []struct {
			name        string
			resource    any
			errContains string
		}{
			{"before-create", &dbtest.TestWithBeforeCreate{}, "gorm callback/hooks are not supported"},
			{"after-create", &dbtest.TestWithAfterCreate{}, "gorm callback/hooks are not supported"},
			{"before-save", &dbtest.TestWithBeforeSave{}, "gorm callback/hooks are not supported"},
			{"before-save", &dbtest.TestWithAfterSave{}, "gorm callback/hooks are not supported"},
			{"nil", nil, "unable to determine type of item"},
		}
		for _, tt := range hookTests {
			t.Run(tt.name, func(t *testing.T) {
				assert, require := assert.New(t), require.New(t)
				w := dbw.New(conn)
				err := w.CreateItems(context.Background(), []any{tt.resource})
				require.Error(err)
				assert.ErrorIs(err, dbw.ErrInvalidParameter)
				assert.Contains(err.Error(), tt.errContains)
			})
		}
	})
}

func TestDb_CreateItems_OnConflict(t *testing.T) {
	ctx := context.Background()
	conn, _ := dbw.TestSetup(t)
	rw := dbw.New(conn)
	dbType, _, err := conn.DbType()
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

	createOnConflictUsers := func(t *testing.T, name string) []*dbtest.TestUser {
		require := require.New(t)
		var conflictUsers []*dbtest.TestUser
		for i := 0; i < 10; i++ {
			initialUser := createInitialUser()
			conflictUser, err := dbtest.NewTestUser()
			require.NoError(err)
			userNameId, err := dbw.NewId(name + strconv.Itoa(i))
			require.NoError(err)
			conflictUser.PublicId = initialUser.PublicId
			conflictUser.Name = userNameId
			conflictUsers = append(conflictUsers, conflictUser)
		}
		return conflictUsers
	}

	tests := []struct {
		name            string
		onConflict      dbw.OnConflict
		setup           func(t *testing.T, name string) []*dbtest.TestUser
		additionalOpts  []dbw.Option
		wantUpdate      bool
		wantEmail       string
		withDebug       bool
		wantErrContains string
	}{
		{
			name: "simple",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			setup:      createOnConflictUsers,
			wantUpdate: true,
		},
		{
			name: "do-nothing",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.DoNothing(true),
			},
			setup:      createOnConflictUsers,
			wantUpdate: false,
		},
		{
			name: "update-all",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.UpdateAll(true),
			},
			setup:      createOnConflictUsers,
			wantUpdate: true,
		},
		{
			name: "on-constraint",
			onConflict: dbw.OnConflict{
				Target: dbw.Constraint("db_test_user_pkey"),
				Action: dbw.SetColumns([]string{"name"}),
			},
			setup:      createOnConflictUsers,
			wantUpdate: true,
		},
		{
			name: "err-vet-for-write",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			setup: func(t *testing.T, name string) []*dbtest.TestUser {
				var conflictUsers []*dbtest.TestUser
				for i := 0; i < 10; i++ {
					conflictUser := dbtest.AllocTestUser()
					conflictUsers = append(conflictUsers, &conflictUser)
				}
				return conflictUsers
			},
			wantErrContains: "dbtest.(TestUser).VetForWrite: missing public id: invalid parameter",
		},
		{
			name: "invalid-conflict-target",
			onConflict: dbw.OnConflict{
				Target: "invalid",
				Action: dbw.SetColumns([]string{"name"}),
			},
			setup:           createOnConflictUsers,
			wantErrContains: "dbw.CreateItems: invalid conflict target string: invalid parameter",
		},
		{
			name: "with-version-success",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			setup:      createOnConflictUsers,
			wantUpdate: true,
			additionalOpts: []dbw.Option{
				dbw.WithVersion(func() *uint32 { i := uint32(1); return &i }()),
			},
		},
		{
			name: "with-version-fail",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumns([]string{"name"}),
			},
			setup:      createOnConflictUsers,
			wantUpdate: false,
			additionalOpts: []dbw.Option{
				dbw.WithVersion(func() *uint32 { i := uint32(10000); return &i }()),
			},
		},
		{
			name: "with-expr-default",
			onConflict: dbw.OnConflict{
				Target: dbw.Columns{"public_id"},
				Action: dbw.SetColumnValues(map[string]any{
					"name":         dbw.Expr("lower(?)", "test with expr and default "),
					"email":        "alice@gmail.com",
					"phone_number": dbw.Expr("NULL"),
				}),
			},
			setup: func(t *testing.T, name string) []*dbtest.TestUser {
				require := require.New(t)
				initialUser := createInitialUser()
				conflictUser, err := dbtest.NewTestUser()
				require.NoError(err)
				conflictUser.PublicId = initialUser.PublicId
				return []*dbtest.TestUser{conflictUser}
			},
			additionalOpts: []dbw.Option{},
			wantUpdate:     true,
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
			var conflictUsers []*dbtest.TestUser
			if tt.setup != nil {
				conflictUsers = tt.setup(t, tt.name)
			}
			var rowsAffected int64
			opts := []dbw.Option{dbw.WithOnConflict(&tt.onConflict), dbw.WithReturnRowsAffected(&rowsAffected)}
			if tt.additionalOpts != nil {
				opts = append(opts, tt.additionalOpts...)
			}
			if tt.withDebug {
				conn.Debug(true)
			}
			err = rw.CreateItems(ctx, conflictUsers, opts...)
			if tt.withDebug {
				conn.Debug(false)
			}
			if tt.wantErrContains != "" {
				require.Error(err)
				assert.Contains(err.Error(), tt.wantErrContains)
				return
			}
			require.NoError(err)
			if tt.wantUpdate {
				assert.GreaterOrEqual(int64(10), rowsAffected)
			} else {
				assert.Equal(int64(0), rowsAffected)
			}
			for _, conflictUser := range conflictUsers {
				foundUser, err := dbtest.NewTestUser()
				require.NoError(err)
				foundUser.PublicId = conflictUser.PublicId
				err = rw.LookupByPublicId(context.Background(), foundUser)
				require.NoError(err)
				t.Log(foundUser)
				if tt.wantUpdate {
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
