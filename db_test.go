package dbw_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
					dbw.WithGormFormatter(hclog.New(hclog.DefaultOptions)),
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
					dbw.WithGormFormatter(hclog.New(hclog.DefaultOptions)),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-sqlite",
			args: args{
				dbType:        dbw.Sqlite,
				connectionUrl: "",
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
		})
	}
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
		tmpDbFile, err := ioutil.TempFile("./", "tmp-db")
		require.NoError(err)
		t.Cleanup(func() {
			os.Remove(tmpDbFile.Name())
			os.Remove(tmpDbFile.Name() + "-journal")
		})
		db, err := dbw.Open(dbw.Sqlite, tmpDbFile.Name())
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
		tmpDbFile, err := ioutil.TempFile("./", "tmp-db")
		require.NoError(err)
		t.Cleanup(func() {
			os.Remove(tmpDbFile.Name())
			os.Remove(tmpDbFile.Name() + "-journal")
		})
		db, err := dbw.Open(dbw.Sqlite, tmpDbFile.Name())
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
