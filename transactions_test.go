package dbw_test

import (
	"context"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRW_Transactions(t *testing.T) {
	t.Parallel()
	testCtx := context.Background()
	conn, _ := dbw.TestSetup(t)

	t.Run("simple", func(t *testing.T) {
		require := require.New(t)
		id, err := dbw.NewId("u")
		require.NoError(err)
		w := dbw.New(conn)

		tx, err := w.Begin(testCtx)
		require.NoError(err)

		user, err := dbtest.NewTestUser()
		require.NoError(err)
		require.NoError(tx.Create(testCtx, &user))

		user.Name = id
		rowsUpdated, err := tx.Update(testCtx, user, []string{"Name"}, nil)
		require.NoError(err)
		require.Equal(1, rowsUpdated)
		require.NoError(tx.Commit(testCtx))
	})
	t.Run("rollback-success", func(t *testing.T) {
		require := require.New(t)
		id, err := dbw.NewId("u")
		require.NoError(err)
		w := dbw.New(conn)

		tx, err := w.Begin(testCtx)
		require.NoError(err)

		user, err := dbtest.NewTestUser()
		require.NoError(err)
		require.NoError(tx.Create(testCtx, &user))

		user.Name = id
		rowsUpdated, err := tx.Update(testCtx, user, []string{"Name"}, nil)
		require.NoError(err)
		require.Equal(1, rowsUpdated)
		require.NoError(tx.Rollback(testCtx))
	})
	t.Run("no-transaction", func(t *testing.T) {
		assert := assert.New(t)
		w := dbw.New(conn)
		assert.Error(w.Rollback(testCtx))
		assert.Error(w.Commit(testCtx))
	})
}
