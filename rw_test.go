package dbw_test

import (
	"context"
	"database/sql"
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
			})

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
		err = w.LookupWhere(context.Background(), &foundUser, "public_id = ?", user.PublicId)
		require.NoError(err)
		assert.Equal(foundUser.PublicId, user.PublicId)
	})
	t.Run("tx-nil,", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.RW{}
		var foundUser dbtest.TestUser
		err := w.LookupWhere(context.Background(), &foundUser, "public_id = ?", 1)
		require.Error(err)
		assert.Equal("dbw.LookupWhere: missing underlying db: invalid parameter", err.Error())
	})
	t.Run("not-found", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), &foundUser, "public_id = ?", id)
		require.Error(err)
		assert.ErrorIs(err, dbw.ErrRecordNotFound)
	})
	t.Run("bad-where", func(t *testing.T) {
		require := require.New(t)
		w := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), &foundUser, "? = ?", id)
		require.Error(err)
	})
	t.Run("not-ptr", func(t *testing.T) {
		require := require.New(t)
		w := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)

		var foundUser dbtest.TestUser
		err = w.LookupWhere(context.Background(), foundUser, "public_id = ?", id)
		require.Error(err)
	})
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
