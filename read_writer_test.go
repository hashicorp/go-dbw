package db_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/hashicorp/go-db"
	"github.com/stretchr/testify/require"
)

func TestDb_Exec(t *testing.T) {
	t.Parallel()
	t.Run("update", func(t *testing.T) {
		testCtx := context.Background()
		conn, _ := db.TestSetup(t)
		db.TestCreateTables(t, conn)
		require := require.New(t)
		w := db.New(conn)
		id := db.TestId(t)
		_, err := w.Exec(testCtx,
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
}
