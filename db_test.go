package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/go-db"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	ctx := context.Background()
	_, url := db.TestSetup(t)

	type args struct {
		dbType        db.DbType
		connectionUrl string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				dbType:        db.Sqlite,
				connectionUrl: url,
			},
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				dbType:        db.Sqlite,
				connectionUrl: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			if tt.args.connectionUrl != "" {
				t.Cleanup(func() { os.Remove(tt.args.connectionUrl + "-journal") })
			}
			got, err := db.Open(tt.args.dbType, tt.args.connectionUrl)
			defer func() {
				if err == nil {
					sqlDB, err := got.SqlDB(ctx)
					require.NoError(err)
					err = sqlDB.Close()
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
