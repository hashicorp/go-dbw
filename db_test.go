// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dbw_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
)

func TestOpen(t *testing.T) {
	ctx := context.Background()
	_, url := dbw.TestSetup(t)

	type args struct {
		dbType        dbw.DbType
		connectionUrl string
		opts          []dbw.Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid-sqlite-with-opts",
			args: args{
				dbType:        dbw.Sqlite,
				connectionUrl: url,
				opts: []dbw.Option{
					dbw.WithMinOpenConnections(1),
					dbw.WithMaxOpenConnections(2),
					dbw.WithLogger(hclog.New(hclog.DefaultOptions)),
				},
			},
			wantErr: false,
		},
		{
			name: "valid-sqlite-no-opts",
			args: args{
				dbType:        dbw.Sqlite,
				connectionUrl: url,
			},
			wantErr: false,
		},
		{
			name: "invalid-connection-opts",
			args: args{
				dbType:        dbw.Sqlite,
				connectionUrl: url,
				opts: []dbw.Option{
					dbw.WithMinOpenConnections(3),
					dbw.WithMaxOpenConnections(2),
					dbw.WithLogger(hclog.New(hclog.DefaultOptions)),
				},
			},
			wantErr: true,
		},
		{
			name: "missing-url",
			args: args{
				dbType:        dbw.Sqlite,
				connectionUrl: "",
			},
			wantErr: true,
		},
		{
			name: "invalid-url",
			args: args{
				dbType:        dbw.Sqlite,
				connectionUrl: "file::memory:?cache=invalid-parameter",
				opts: []dbw.Option{dbw.WithLogger(hclog.New(
					&hclog.LoggerOptions{
						Output: ioutil.Discard,
					},
				))},
			},
			wantErr: true,
		},
		{
			name: "unknown-type",
			args: args{
				dbType:        dbw.UnknownDB,
				connectionUrl: url,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			t.Cleanup(func() {
				os.Remove(tt.args.connectionUrl + "-journal")
				os.Remove(tt.args.connectionUrl)
			})

			got, err := dbw.Open(tt.args.dbType, tt.args.connectionUrl, tt.args.opts...)
			defer func() {
				if err == nil {
					err = got.Close(ctx)
					require.NoError(err)
				}
			}()
			if tt.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			rw := dbw.New(got)
			rows, err := rw.Query(context.Background(), "PRAGMA foreign_keys", nil)
			require.NoError(err)
			require.True(rows.Next())
			type foo struct{}
			f := struct {
				ForeignKeys int
			}{}
			err = rw.ScanRows(rows, &f)
			require.NoError(err)
			require.Equal(1, f.ForeignKeys)
			fmt.Println(f)
		})
	}
}

func TestDB_OpenWith(t *testing.T) {
	t.Run("simple-sqlite", func(t *testing.T) {
		assert := assert.New(t)
		_, err := dbw.OpenWith(sqlite.Open("file::memory:"), nil)
		assert.NoError(err)
	})
}

func TestDB_StringToDbType(t *testing.T) {
	tests := []struct {
		name    string
		want    dbw.DbType
		wantErr bool
	}{
		{name: "postgres", want: dbw.Postgres},
		{name: "sqlite", want: dbw.Sqlite},
		{name: "unknown", want: dbw.UnknownDB, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			got, err := dbw.StringToDbType(tt.name)
			if tt.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(got, tt.want)
		})
	}
}

func TestDB_SqlDB(t *testing.T) {
	testCtx := context.Background()
	t.Run("valid", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db, err := dbw.Open(dbw.Sqlite, "file::memory:")
		require.NoError(err)
		got, err := db.SqlDB(testCtx)
		require.NoError(err)
		assert.NotNil(got)
	})

	t.Run("invalid", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db := &dbw.DB{}
		got, err := db.SqlDB(testCtx)
		require.Error(err)
		assert.Nil(got)
	})
}

func TestDB_Close(t *testing.T) {
	testCtx := context.Background()
	t.Run("valid", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db, err := dbw.Open(dbw.Sqlite, "file::memory:")
		require.NoError(err)
		got, err := db.SqlDB(testCtx)
		require.NoError(err)
		require.NotNil(got)
		assert.NoError(got.Close())
	})
	t.Run("invalid", func(t *testing.T) {
		assert := assert.New(t)
		db := &dbw.DB{}
		err := db.Close(testCtx)
		assert.Error(err)
	})
}

func TestDB_LogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level dbw.LogLevel
	}{
		{"default", dbw.Default},
		{"silent", dbw.Silent},
		{"error", dbw.Error},
		{"warn", dbw.Warn},
		{"info", dbw.Info},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := dbw.TestSetup(t)
			db.LogLevel(tt.level)
		})
	}
}
