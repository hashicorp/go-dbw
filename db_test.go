package dbw_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	ctx := context.Background()
	_, url := dbw.TestSetup(t)

	type args struct {
		dbType        dbw.DbType
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
				dbType:        dbw.Sqlite,
				connectionUrl: url,
			},
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				dbType:        dbw.Sqlite,
				connectionUrl: "",
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

			got, err := dbw.Open(tt.args.dbType, tt.args.connectionUrl)
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
