package dbw_test

import (
	"context"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDb_LookupBy(t *testing.T) {
	t.Parallel()
	db, _ := dbw.TestSetup(t)
	testRw := dbw.New(db)
	scooter := testScooter(t, testRw, "", 0)
	user := testUser(t, testRw, "", "", "")
	car := testCar(t, testRw)
	rental := testRental(t, testRw, user.PublicId, car.PublicId)
	// scooterAccessory := testScooterAccessory(t, db, scooter.Id, accessory.AccessoryId)

	type args struct {
		resource interface{}
		opt      []dbw.Option
	}
	tests := []struct {
		name      string
		rw        *dbw.RW
		args      args
		wantErr   bool
		want      proto.Message
		wantIsErr error
	}{
		{
			name: "simple-private-id",
			rw:   testRw,
			args: args{
				resource: scooter,
			},
			wantErr: false,
			want:    scooter,
		},
		{
			name: "simple-public-id",
			rw:   testRw,
			args: args{
				resource: user,
			},
			wantErr: false,
			want:    user,
		},
		{
			name: "with-table",
			rw:   testRw,
			args: args{
				resource: user,
				opt:      []dbw.Option{dbw.WithTable(user.TableName())},
			},
			wantErr: false,
			want:    user,
		},
		{
			name: "with-debug",
			rw:   testRw,
			args: args{
				resource: user,
				opt:      []dbw.Option{dbw.WithDebug(true)},
			},
			wantErr: false,
			want:    user,
		},
		{
			name: "with-table-fail",
			rw:   testRw,
			args: args{
				resource: user,
				opt:      []dbw.Option{dbw.WithTable("invalid-table-name")},
			},
			wantErr: true,
		},
		{
			name: "compond",
			rw:   testRw,
			args: args{
				resource: rental,
			},
			wantErr: false,
			want:    rental,
		},
		{
			name: "compond-with-zero-value-pk",
			rw:   testRw,
			args: args{
				resource: func() interface{} {
					cp := rental.Clone()
					cp.(*dbtest.TestRental).CarId = ""
					return cp
				}(),
			},
			wantErr:   true,
			wantIsErr: dbw.ErrInvalidParameter,
		},
		{
			name: "missing-public-id",
			rw:   testRw,
			args: args{
				resource: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{},
				},
			},
			wantErr:   true,
			wantIsErr: dbw.ErrInvalidParameter,
		},
		{
			name: "missing-private-id",
			rw:   testRw,
			args: args{
				resource: &dbtest.TestScooter{
					StoreTestScooter: &dbtest.StoreTestScooter{},
				},
			},
			wantErr:   true,
			wantIsErr: dbw.ErrInvalidParameter,
		},
		{
			name: "not-an-ider",
			rw:   testRw,
			args: args{
				resource: &dbtest.NotIder{},
			},
			wantErr:   true,
			wantIsErr: dbw.ErrInvalidParameter,
		},
		{
			name: "missing-underlying-db",
			rw:   &dbw.RW{},
			args: args{
				resource: user,
			},
			wantErr:   true,
			wantIsErr: dbw.ErrInvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			cloner, ok := tt.args.resource.(dbtest.Cloner)
			require.True(ok)
			cp := cloner.Clone()
			err := tt.rw.LookupBy(context.Background(), cp, tt.args.opt...)
			if tt.wantErr {
				require.Error(err)
				if tt.wantIsErr != nil {
					assert.ErrorIs(err, tt.wantIsErr)
				}
				return
			}
			require.NoError(err)
			assert.True(proto.Equal(tt.want, cp.(proto.Message)))
		})
	}
	t.Run("not-ptr", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		u := testUser(t, testRw, "", "", "")
		err := testRw.LookupBy(context.Background(), *u)
		require.Error(err)
		assert.ErrorIs(err, dbw.ErrInvalidParameter)
	})
	t.Run("hooks", func(t *testing.T) {
		hookTests := []struct {
			name     string
			resource interface{}
		}{
			{"after", &dbtest.TestWithAfterFind{}},
		}
		for _, tt := range hookTests {
			t.Run(tt.name, func(t *testing.T) {
				assert, require := assert.New(t), require.New(t)
				w := dbw.New(db)
				err := w.LookupBy(context.Background(), tt.resource)
				require.Error(err)
				assert.ErrorIs(err, dbw.ErrInvalidParameter)
				assert.Contains(err.Error(), "gorm callback/hooks are not supported")
			})
		}
	})
}
