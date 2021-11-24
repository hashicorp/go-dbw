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

func TestDb_Query(t *testing.T) {
	const (
		insert = "insert into db_test_user (public_id, name) values(@public_id, @name)"
		query  = "select * from db_test_user where name in (?, ?)"
	)
	testCtx := context.Background()
	t.Parallel()
	conn, _ := dbw.TestSetup(t)
	t.Run("valid", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.New(conn)
		publicId, err := dbw.NewPublicId("u")
		require.NoError(err)
		rowsAffected, err := rw.Exec(testCtx, insert, []interface{}{
			sql.Named("public_id", publicId),
			sql.Named("name", "alice"),
		})
		require.NoError(err)
		require.Equal(1, rowsAffected)
		rows, err := rw.Query(testCtx, query, []interface{}{"alice", "bob"})
		require.NoError(err)
		defer func() { err := rows.Close(); assert.NoError(err) }()
		for rows.Next() {
			u, err := dbtest.NewTestUser()
			require.NoError(err)
			// scan the row into your struct
			err = rw.ScanRows(rows, &u)
			require.NoError(err)
			assert.Equal(publicId, u.PublicId)
		}
	})
	t.Run("missing-sql", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.New(conn)
		got, err := rw.Query(testCtx, "", nil)
		require.Error(err)
		assert.Zero(got)
	})
	t.Run("missing-underlying-db", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.RW{}
		got, err := rw.Query(testCtx, "", nil)
		require.Error(err)
		assert.Zero(got)
	})
	t.Run("bad-sql", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.New(conn)
		got, err := rw.Query(testCtx, "from", nil)
		require.Error(err)
		assert.Zero(got)
	})
}
