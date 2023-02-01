// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	t.Parallel()
	const (
		insert = "insert into db_test_user (public_id, name) values(@public_id, @name)"
		query  = "select * from db_test_user where name in (?, ?)"
	)
	testCtx := context.Background()
	conn, _ := dbw.TestSetup(t)
	t.Run("valid", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.New(conn)
		publicId, err := dbw.NewId("u")
		require.NoError(err)
		rowsAffected, err := rw.Exec(testCtx, insert, []interface{}{
			sql.Named("public_id", publicId),
			sql.Named("name", "alice"),
		})
		require.NoError(err)
		require.Equal(1, rowsAffected)
		rows, err := rw.Query(testCtx, query, []interface{}{"alice", "bob"}, dbw.WithDebug(true))
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

func TestDb_ScanRows(t *testing.T) {
	t.Parallel()
	testCtx := context.Background()
	conn, _ := dbw.TestSetup(t)
	rw := dbw.New(conn)
	t.Run("valid", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		user, err := dbtest.NewTestUser()
		require.NoError(err)
		err = rw.Create(testCtx, user)
		require.NoError(err)
		assert.NotEmpty(user.PublicId)
		where := "select * from db_test_user where name in (?, ?)"
		rows, err := rw.Query(context.Background(), where, []interface{}{"alice", "bob"})
		require.NoError(err)
		defer func() { err := rows.Close(); assert.NoError(err) }()
		for rows.Next() {
			u := dbtest.AllocTestUser()
			// scan the row into your struct
			err = rw.ScanRows(rows, &u)
			require.NoError(err)
			assert.Equal(user.PublicId, u.PublicId)
		}
	})
	t.Run("missing-underlying-db", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		rw := dbw.RW{}
		u := dbtest.AllocTestUser()
		err := rw.ScanRows(&sql.Rows{}, &u)
		require.Error(err)
		assert.Contains(err.Error(), "missing underlying db")
	})
	t.Run("missing-result", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		err := rw.ScanRows(&sql.Rows{}, nil)
		require.Error(err)
		assert.Contains(err.Error(), "missing result")
	})
	t.Run("missing-rows", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		u := dbtest.AllocTestUser()
		err := rw.ScanRows(nil, &u)
		require.Error(err)
		assert.Contains(err.Error(), "missing rows")
	})
}
