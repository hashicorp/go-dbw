package dbw_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDb_Create(t *testing.T) {
	testCtx := context.Background()
	db, _ := dbw.TestSetup(t)
	t.Run("simple", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		id, err := uuid.GenerateUUID()
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.PublicId = id
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
	t.Run("WithBeforeCreate", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(db)
		id, err := uuid.GenerateUUID()
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.PublicId = id
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
		id, err := uuid.GenerateUUID()
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.PublicId = id
		user.Name = "alice" + id
		fn := func(i interface{}) error {
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

		fn = func(i interface{}) error {
			return errors.New("fail")
		}
		err = w.Create(
			context.Background(),
			user,
			dbw.WithAfterWrite(fn),
		)
		require.Error(err)

	})
	t.Run("nil-tx", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(nil)
		id, err := uuid.GenerateUUID()
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "foo-" + id
		err = w.Create(context.Background(), user)
		require.Error(err)
		assert.Contains(err.Error(), "db.Create: missing underlying db: invalid parameter")
	})
}
