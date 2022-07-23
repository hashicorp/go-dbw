package dbw_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDb_DoTx(t *testing.T) {
	t.Parallel()
	testCtx := context.TODO()
	conn, _ := dbw.TestSetup(t)
	retryErr := errors.New("retry error")
	retryOnFn := func(err error) bool {
		if errors.Is(err, retryErr) {
			return true
		}
		return false
	}
	t.Run("timed-out", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		timeoutCtx, timeoutCancel := context.WithTimeout(testCtx, 1*time.Microsecond)
		defer timeoutCancel()

		w := dbw.New(conn)
		attempts := 0
		_, err := w.DoTx(timeoutCtx, retryOnFn, 2, dbw.ConstBackoff{DurationMs: 1}, func(dbw.Reader, dbw.Writer) error {
			attempts += 1
			return retryErr
		})
		require.Error(err)
		assert.Contains(err.Error(), "dbw.DoTx: context deadline exceeded")
	})
	t.Run("valid-with-10-retries", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 10, dbw.ExpBackoff{},
			func(dbw.Reader, dbw.Writer) error {
				attempts += 1
				if attempts < 9 {
					return retryErr
				}
				return nil
			})
		require.NoError(err)
		assert.Equal(8, got.Retries)
		assert.Equal(9, attempts) // attempted 1 + 8 retries
	})
	t.Run("valid-with-1-retries", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 1, dbw.ExpBackoff{},
			func(dbw.Reader, dbw.Writer) error {
				attempts += 1
				if attempts < 2 {
					return retryErr
				}
				return nil
			})
		require.NoError(err)
		assert.Equal(1, got.Retries)
		assert.Equal(2, attempts) // attempted 1 + 8 retries
	})
	t.Run("valid-with-2-retries", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 3, dbw.ExpBackoff{},
			func(dbw.Reader, dbw.Writer) error {
				attempts += 1
				if attempts < 3 {
					return retryErr
				}
				return nil
			})
		require.NoError(err)
		assert.Equal(2, got.Retries)
		assert.Equal(3, attempts) // attempted 1 + 8 retries
	})
	t.Run("valid-with-4-retries", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 4, dbw.ExpBackoff{},
			func(dbw.Reader, dbw.Writer) error {
				attempts += 1
				if attempts < 4 {
					return retryErr
				}
				return nil
			})
		require.NoError(err)
		assert.Equal(3, got.Retries)
		assert.Equal(4, attempts) // attempted 1 + 8 retries
	})
	t.Run("zero-retries", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 0, dbw.ExpBackoff{}, func(dbw.Reader, dbw.Writer) error { attempts += 1; return nil })
		require.NoError(err)
		assert.Equal(dbw.RetryInfo{}, got)
		assert.Equal(1, attempts)
	})
	t.Run("nil-tx", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := &dbw.RW{}
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 1, dbw.ExpBackoff{}, func(dbw.Reader, dbw.Writer) error { attempts += 1; return nil })
		require.Error(err)
		assert.Equal(dbw.RetryInfo{}, got)
		assert.Equal("dbw.DoTx: missing underlying db: invalid parameter", err.Error())
	})
	t.Run("nil-retryOnFn", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, nil, 1, dbw.ExpBackoff{}, func(dbw.Reader, dbw.Writer) error { attempts += 1; return nil })
		require.Error(err)
		assert.Equal(dbw.RetryInfo{}, got)
		assert.Equal("dbw.DoTx: missing retry errors matching function: invalid parameter", err.Error())
	})
	t.Run("nil-handler", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		got, err := w.DoTx(testCtx, retryOnFn, 1, dbw.ExpBackoff{}, nil)
		require.Error(err)
		assert.Equal(dbw.RetryInfo{}, got)
		assert.Equal("dbw.DoTx: missing handler: invalid parameter", err.Error())
	})
	t.Run("nil-backoff", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 1, nil, func(dbw.Reader, dbw.Writer) error { attempts += 1; return nil })
		require.Error(err)
		assert.Equal(dbw.RetryInfo{}, got)
		assert.Equal("dbw.DoTx: missing backoff: invalid parameter", err.Error())
	})
	t.Run("not-a-retry-err", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		got, err := w.DoTx(testCtx, retryOnFn, 1, dbw.ExpBackoff{}, func(dbw.Reader, dbw.Writer) error { return errors.New("not a retry error") })
		require.Error(err)
		assert.Equal(dbw.RetryInfo{}, got)
		assert.False(errors.Is(err, retryErr))
	})
	t.Run("too-many-retries", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		attempts := 0
		got, err := w.DoTx(testCtx, retryOnFn, 2, dbw.ConstBackoff{}, func(dbw.Reader, dbw.Writer) error {
			attempts += 1
			return retryErr
		})
		require.Error(err)
		assert.Equal(3, got.Retries)
		assert.Contains(err.Error(), "dbw.DoTx: too many retries: 3 of 3")
	})
	t.Run("updating-good-bad-good", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.New(conn)
		id, err := dbw.NewId("i")
		require.NoError(err)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		user.Name = "foo-" + id
		err = rw.Create(context.Background(), user)
		require.NoError(err)

		_, err = rw.DoTx(testCtx, retryOnFn, 10, dbw.ExpBackoff{}, func(r dbw.Reader, w dbw.Writer) error {
			user.Name = "friendly-" + id
			rowsUpdated, err := w.Update(context.Background(), user, []string{"Name"}, nil)
			if err != nil {
				return err
			}
			if rowsUpdated != 1 {
				return fmt.Errorf("error in number of rows updated %d", rowsUpdated)
			}
			return nil
		})
		require.NoError(err)

		foundUser := dbtest.AllocTestUser()
		assert.NoError(err)
		foundUser.PublicId = user.PublicId
		err = rw.LookupByPublicId(context.Background(), &foundUser)
		require.NoError(err)
		assert.Equal(foundUser.Name, user.Name)

		user2, err := dbtest.NewTestUser()
		require.NoError(err)
		_, err = rw.DoTx(testCtx, retryOnFn, 10, dbw.ExpBackoff{}, func(_ dbw.Reader, w dbw.Writer) error {
			user2.Name = "friendly2-" + id
			rowsUpdated, err := w.Update(context.Background(), user2, []string{"Name"}, nil)
			if err != nil {
				return err
			}
			if rowsUpdated != 1 {
				return fmt.Errorf("error in number of rows updated %d", rowsUpdated)
			}
			return nil
		})
		require.Error(err)
		err = rw.LookupByPublicId(context.Background(), &foundUser)
		require.NoError(err)
		assert.NotEqual(foundUser.Name, user2.Name)

		_, err = rw.DoTx(testCtx, retryOnFn, 10, dbw.ExpBackoff{}, func(r dbw.Reader, w dbw.Writer) error {
			user.Name = "friendly2-" + id
			rowsUpdated, err := w.Update(context.Background(), user, []string{"Name"}, nil)
			if err != nil {
				return err
			}
			if rowsUpdated != 1 {
				return fmt.Errorf("error in number of rows updated %d", rowsUpdated)
			}
			return nil
		})
		require.NoError(err)
		err = rw.LookupByPublicId(context.Background(), &foundUser)
		require.NoError(err)
		assert.Equal(foundUser.Name, user.Name)
	})
}
