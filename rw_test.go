// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dbw_test

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDb_Exec(t *testing.T) {
	t.Parallel()
	testCtx := context.Background()
	conn, _ := dbw.TestSetup(t)
	t.Run("update", func(t *testing.T) {
		require := require.New(t)
		w := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)
		_, err = w.Exec(testCtx,
			"insert into db_test_user(public_id, name) values(@public_id, @name)",
			[]interface{}{
				sql.Named("public_id", id),
				sql.Named("name", "alice"),
			},
			dbw.WithDebug(true),
		)

		require.NoError(err)
		rowsAffected, err := w.Exec(testCtx,
			"update db_test_user set name = @name where public_id = @public_id",
			[]interface{}{
				sql.Named("public_id", id),
				sql.Named("name", "alice-"+id),
			})
		require.NoError(err)
		require.Equal(1, rowsAffected)
	})
	t.Run("missing-sql", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.New(conn)
		got, err := rw.Exec(testCtx, "", nil)
		require.Error(err)
		assert.Zero(got)
	})
	t.Run("missing-underlying-db", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.RW{}
		got, err := rw.Exec(testCtx, "", nil)
		require.Error(err)
		assert.Zero(got)
	})
	t.Run("bad-sql", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.New(conn)
		got, err := rw.Exec(testCtx, "insert from", nil)
		require.Error(err)
		assert.Zero(got)
	})
}

func TestDb_LookupWhere(t *testing.T) {
	t.Parallel()
	conn, _ := dbw.TestSetup(t)
	t.Run("simple", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "foo-" + user.PublicId
		err = w.Create(context.Background(), user)
		require.NoError(err)
		assert.NotEmpty(user.PublicId)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), &foundUser, "public_id = ? and 1 = ?", []interface{}{user.PublicId, 1}, dbw.WithDebug(true))
		require.NoError(err)
		assert.Equal(foundUser.PublicId, user.PublicId)
	})
	t.Run("with-table", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "foo-" + user.PublicId
		err = w.Create(context.Background(), user, dbw.WithTable(user.TableName()))
		require.NoError(err)
		assert.NotEmpty(user.PublicId)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), &foundUser, "public_id = ?", []interface{}{user.PublicId}, dbw.WithTable(user.TableName()))
		require.NoError(err)
		assert.Equal(foundUser.PublicId, user.PublicId)

		err = w.LookupWhere(context.Background(), &foundUser, "public_id = ?", []interface{}{user.PublicId}, dbw.WithTable("invalid-table-name"))
		require.Error(err)
	})
	t.Run("tx-nil,", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.RW{}
		var foundUser dbtest.TestUser
		err := w.LookupWhere(context.Background(), &foundUser, "public_id = ?", []interface{}{1})
		require.Error(err)
		assert.Equal("dbw.LookupWhere: missing underlying db: invalid parameter", err.Error())
	})
	t.Run("not-found", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), &foundUser, "public_id = ?", []interface{}{id})
		require.Error(err)
		assert.ErrorIs(err, dbw.ErrRecordNotFound)
	})
	t.Run("bad-where", func(t *testing.T) {
		require := require.New(t)
		w := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), &foundUser, "? = ?", []interface{}{id})
		require.Error(err)
	})
	t.Run("not-ptr", func(t *testing.T) {
		require := require.New(t)
		w := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), foundUser, "public_id = ?", []interface{}{id})
		require.Error(err)
	})
	t.Run("hooks", func(t *testing.T) {
		hookTests := []struct {
			name     string
			resource interface{}
		}{
			{"after", &dbtest.TestWithAfterFind{}},
		}
		for _, tt := range hookTests {
			t.Run(tt.name, func(t *testing.T) {
				assert, require := assert.New(t), require.New(t)
				w := dbw.New(conn)
				err := w.LookupWhere(context.Background(), tt.resource, "public_id = ?", []interface{}{"1"})
				require.Error(err)
				assert.ErrorIs(err, dbw.ErrInvalidParameter)
				assert.Contains(err.Error(), "gorm callback/hooks are not supported")
			})
		}
	})
}

func TestDb_SearchWhere(t *testing.T) {
	t.Parallel()
	conn, _ := dbw.TestSetup(t)
	testRw := dbw.New(conn)
	knownUser := testUser(t, testRw, "zedUser", "", "")

	type args struct {
		where string
		arg   []interface{}
		opt   []dbw.Option
	}
	tests := []struct {
		name          string
		rw            *dbw.RW
		createCnt     int
		args          args
		wantCnt       int
		wantErr       bool
		wantNameOrder bool
	}{
		{
			name:      "no-limit",
			rw:        testRw,
			createCnt: 10,
			args: args{
				where: "1=1",
				opt:   []dbw.Option{dbw.WithLimit(-1), dbw.WithOrder("name asc")},
			},
			wantCnt:       11, // there's an additional knownUser
			wantErr:       false,
			wantNameOrder: true,
		},
		{
			name:      "no-where",
			rw:        testRw,
			createCnt: 10,
			args: args{
				opt: []dbw.Option{dbw.WithLimit(10)},
			},
			wantCnt: 10,
			wantErr: false,
		},
		{
			name:      "custom-limit",
			rw:        testRw,
			createCnt: 10,
			args: args{
				where: "1=1",
				opt:   []dbw.Option{dbw.WithLimit(3)},
			},
			wantCnt: 3,
			wantErr: false,
		},
		{
			name:      "simple",
			rw:        testRw,
			createCnt: 1,
			args: args{
				where: "public_id = ?",
				arg:   []interface{}{knownUser.PublicId},
				opt:   []dbw.Option{dbw.WithLimit(3)},
			},
			wantCnt: 1,
			wantErr: false,
		},
		{
			name:      "with-table",
			rw:        testRw,
			createCnt: 1,
			args: args{
				where: "public_id = ?",
				arg:   []interface{}{knownUser.PublicId},
				opt:   []dbw.Option{dbw.WithLimit(3), dbw.WithTable(knownUser.TableName())},
			},
			wantCnt: 1,
			wantErr: false,
		},
		{
			name:      "with-table-fail",
			rw:        testRw,
			createCnt: 1,
			args: args{
				where: "public_id = ?",
				arg:   []interface{}{knownUser.PublicId},
				opt:   []dbw.Option{dbw.WithLimit(3), dbw.WithTable("invalid-table-name")},
			},
			wantErr: true,
		},
		{
			name:      "no args",
			rw:        testRw,
			createCnt: 1,
			args: args{
				where: fmt.Sprintf("public_id = '%v'", knownUser.PublicId),
				opt:   []dbw.Option{dbw.WithLimit(3)},
			},
			wantCnt: 1,
			wantErr: false,
		},
		{
			name:      "no where, but with args",
			rw:        testRw,
			createCnt: 1,
			args: args{
				arg: []interface{}{knownUser.PublicId},
				opt: []dbw.Option{dbw.WithLimit(3)},
			},
			wantErr: true,
		},
		{
			name:      "not-found",
			rw:        testRw,
			createCnt: 1,
			args: args{
				where: "public_id = ?",
				arg:   []interface{}{"bad-id"},
				opt:   []dbw.Option{dbw.WithLimit(3)},
			},
			wantCnt: 0,
			wantErr: false,
		},
		{
			name:      "bad-where",
			rw:        testRw,
			createCnt: 1,
			args: args{
				where: "bad_column_name = ?",
				arg:   []interface{}{knownUser.PublicId},
				opt:   []dbw.Option{dbw.WithLimit(3)},
			},
			wantCnt: 0,
			wantErr: true,
		},
		{
			name:      "nil-underlying",
			rw:        &dbw.RW{},
			createCnt: 1,
			args: args{
				where: "public_id = ?",
				arg:   []interface{}{knownUser.PublicId},
				opt:   []dbw.Option{dbw.WithLimit(3)},
			},
			wantCnt: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			testUsers := []*dbtest.TestUser{}
			for i := 0; i < tt.createCnt; i++ {
				testUsers = append(testUsers, testUser(t, testRw, tt.name+strconv.Itoa(i), "", ""))
			}
			assert.Equal(tt.createCnt, len(testUsers))

			var foundUsers []*dbtest.TestUser
			err := tt.rw.SearchWhere(context.Background(), &foundUsers, tt.args.where, tt.args.arg, tt.args.opt...)
			if tt.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tt.wantCnt, len(foundUsers))
			if tt.wantNameOrder {
				assert.Equal(tt.name+strconv.Itoa(0), foundUsers[0].Name)
				for i, u := range foundUsers {
					if u.Name != "zedUser" {
						assert.Equal(tt.name+strconv.Itoa(i), u.Name)
					}
				}
			}
		})
	}
	t.Run("hooks", func(t *testing.T) {
		hookTests := []struct {
			name     string
			resource interface{}
		}{
			{"after", &dbtest.TestWithAfterFind{}},
		}
		for _, tt := range hookTests {
			t.Run(tt.name, func(t *testing.T) {
				assert, require := assert.New(t), require.New(t)
				w := dbw.New(conn)
				err := w.SearchWhere(context.Background(), tt.resource, "public_id = 1", nil)
				require.Error(err)
				assert.ErrorIs(err, dbw.ErrInvalidParameter)
				assert.Contains(err.Error(), "gorm callback/hooks are not supported")
			})
		}
	})
}

func TestRW_IsTx(t *testing.T) {
	t.Parallel()
	testCtx := context.Background()
	conn, _ := dbw.TestSetup(t)
	testRw := dbw.New(conn)
	assert, require := assert.New(t), require.New(t)

	assert.False(testRw.IsTx())

	tx, err := testRw.Begin(testCtx)
	require.NoError(err)
	assert.NotNil(tx)
	assert.True(tx.IsTx())
}

func TestDialect(t *testing.T) {
	t.Parallel()
	conn, _ := dbw.TestSetup(t)
	testRw := dbw.New(conn)
	assert, require := assert.New(t), require.New(t)

	gotTyp, gotRawName, err := testRw.Dialect()
	require.NoError(err)
	typ, rawName, err := conn.DbType()
	require.NoError(err)
	assert.Equal(typ, gotTyp)
	assert.Equal(rawName, gotRawName)
}

func testUser(t *testing.T, rw *dbw.RW, name, email, phoneNumber string) *dbtest.TestUser {
	t.Helper()
	require := require.New(t)
	r, err := dbtest.NewTestUser()
	require.NoError(err)
	r.Name = name
	r.Email = email
	r.PhoneNumber = phoneNumber
	if rw != nil {
		err = rw.Create(context.Background(), r)
		require.NoError(err)
	}
	return r
}

func testCar(t *testing.T, rw *dbw.RW) *dbtest.TestCar {
	t.Helper()
	require := require.New(t)
	r, err := dbtest.NewTestCar()
	require.NoError(err)
	if rw != nil {
		err = rw.Create(context.Background(), r)
		require.NoError(err)
	}
	return r
}

func testScooter(t *testing.T, rw *dbw.RW, model string, mpg int32) *dbtest.TestScooter {
	t.Helper()
	require := require.New(t)
	r, err := dbtest.NewTestScooter()
	require.NoError(err)
	r.Model = model
	r.Mpg = mpg
	if rw != nil {
		err = rw.Create(context.Background(), r)
		require.NoError(err)
	}
	return r
}

func testRental(t *testing.T, rw *dbw.RW, userId, carId string) *dbtest.TestRental {
	t.Helper()
	require := require.New(t)
	r, err := dbtest.NewTestRental(userId, carId)
	require.NoError(err)
	if rw != nil {
		err = rw.Create(context.Background(), r)
		require.NoError(err)
	}
	return r
}
