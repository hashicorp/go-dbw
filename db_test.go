package db

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	ctx := context.Background()
	_, url := TestSetup(t, "sqlite")

	type args struct {
		dbType        DbType
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
				dbType:        Sqlite,
				connectionUrl: url,
			},
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				dbType:        Sqlite,
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
			got, err := Open(tt.args.dbType, tt.args.connectionUrl)
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
