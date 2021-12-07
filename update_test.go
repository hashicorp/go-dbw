package dbw_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDb_UpdateUnsetField(t *testing.T) {
	t.Parallel()
	testCtx := context.Background()
	assert, require := assert.New(t), require.New(t)
	conn, _ := dbw.TestSetup(t)
	rw := dbw.New(conn)
	tu, err := dbtest.NewTestUser()
	tu.Name = "default"
	require.NoError(err)

	require.NoError(rw.Create(testCtx, tu))

	updatedTu := tu.Clone().(*dbtest.TestUser)
	updatedTu.Name = "updated"
	updatedTu.Email = "ignore"
	cnt, err := rw.Update(testCtx, updatedTu, []string{"Name"}, nil)
	require.NoError(err)
	assert.Equal(1, cnt)
	assert.Equal("", updatedTu.Email)
	assert.Equal("updated", updatedTu.Name)
}

func TestDb_Update(t *testing.T) {
	conn, _ := dbw.TestSetup(t)
	now := &dbtest.Timestamp{Timestamp: timestamppb.Now()}
	publicId, err := dbw.NewId("testuser")
	require.NoError(t, err)
	id, err := dbw.NewId("i")
	require.NoError(t, err)

	badVersion := uint32(22)
	versionOne := uint32(1)
	versionZero := uint32(0)

	successBeforeFn := func(_ interface{}) error {
		return nil
	}
	successAfterFn := func(_ interface{}, _ int) error {
		return nil
	}
	errFailedFn := errors.New("fail")
	failedBeforeFn := func(_ interface{}) error {
		return errFailedFn
	}
	failedAfterFn := func(_ interface{}, _ int) error {
		return errFailedFn
	}

	type args struct {
		i              *dbtest.TestUser
		fieldMaskPaths []string
		setToNullPaths []string
		opt            []dbw.Option
	}
	tests := []struct {
		name            string
		args            args
		want            int
		wantErr         bool
		wantErrMsg      string
		wantName        string
		wantEmail       string
		wantPhoneNumber string
		wantVersion     int
	}{
		{
			name: "simple",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "simple-updated" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
			},
			want:            1,
			wantErr:         false,
			wantErrMsg:      "",
			wantName:        "simple-updated" + id,
			wantEmail:       "",
			wantPhoneNumber: "updated" + id,
		},
		{
			name: "simple-with-bad-version",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "simple-with-bad-version" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithVersion(&badVersion)},
			},
			want:       0,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "simple-with-zero-version",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "simple-with-bad-version" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithVersion(&versionZero)},
			},
			want:       0,
			wantErr:    true,
			wantErrMsg: "with version option is zero: invalid parameter",
		},
		{
			name: "simple-with-version",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "simple-with-version" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithVersion(&versionOne)},
			},
			want:            1,
			wantErr:         false,
			wantErrMsg:      "",
			wantName:        "simple-with-version" + id,
			wantEmail:       "",
			wantPhoneNumber: "updated" + id,
			wantVersion:     2,
		},
		{
			name: "simple-with-where-not-found",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "simple-with-where" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithWhere("name = @name", sql.Named("name", "not-matching"))},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "simple-with-where",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "simple-with-where" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithWhere("email = ? and phone_number = ?", id, id)},
			},
			want:            1,
			wantErr:         false,
			wantErrMsg:      "",
			wantName:        "simple-with-where" + id,
			wantEmail:       "",
			wantPhoneNumber: "updated" + id,
			wantVersion:     2,
		},
		{
			name: "simple-with-where-and-version",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "simple-with-where-and-version" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithWhere("email = ? and phone_number = ?", id, id), dbw.WithVersion(&versionOne)},
			},
			want:            1,
			wantErr:         false,
			wantErrMsg:      "",
			wantName:        "simple-with-where-and-version" + id,
			wantEmail:       "",
			wantPhoneNumber: "updated" + id,
			wantVersion:     2,
		},
		{
			name: "bad-with-where",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "bad-with-where" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithWhere("foo = ? and phone_number = ?", id, id)},
			},
			want:       0,
			wantErr:    true,
			wantErrMsg: `column`,
		},
		{
			name: "multiple-null",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "multiple-null-updated" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name"},
				setToNullPaths: []string{"Email", "PhoneNumber"},
			},
			want:            1,
			wantErr:         false,
			wantErrMsg:      "",
			wantName:        "multiple-null-updated" + id,
			wantEmail:       "",
			wantPhoneNumber: "",
		},
		{
			name: "non-updatable",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "non-updatable" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
						PublicId:    publicId,
						CreateTime:  now,
						UpdateTime:  now,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber", "CreateTime", "UpdateTime", "PublicId"},
				setToNullPaths: []string{"Email"},
			},
			want:            1,
			wantErr:         false,
			wantErrMsg:      "",
			wantName:        "non-updatable" + id,
			wantEmail:       "",
			wantPhoneNumber: "updated" + id,
		},
		{
			name: "both are missing",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "both are missing-updated" + id,
						Email:       id,
						PhoneNumber: id,
					},
				},
				fieldMaskPaths: nil,
				setToNullPaths: []string{},
			},
			want:       0,
			wantErr:    true,
			wantErrMsg: "dbw.Update: both fieldMaskPaths and setToNullPaths are missing: invalid parameter",
		},
		{
			name: "i is nil",
			args: args{
				i:              nil,
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
			},
			want:       0,
			wantErr:    true,
			wantErrMsg: "dbw.Update: missing interface: invalid parameter",
		},
		{
			name: "only read-only",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "only read-only" + id,
						Email:       id,
						PhoneNumber: id,
					},
				},
				fieldMaskPaths: []string{"CreateTime"},
				setToNullPaths: []string{"UpdateTime"},
			},
			want:       0,
			wantErr:    true,
			wantErrMsg: "dbw.Update: after filtering non-updated fields, there are no fields left in fieldMaskPaths or setToNullPaths: invalid parameter",
		},
		{
			name: "intersection",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "alice" + id,
						Email:       id,
						PhoneNumber: id,
					},
				},
				fieldMaskPaths: []string{"Name"},
				setToNullPaths: []string{"Name"},
			},
			want:       0,
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: fieldMashPaths and setToNullPaths cannot intersect: invalid parameter",
		},
		{
			name: "with-before-after-write-success",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "with-before-after-write-success" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithBeforeWrite(successBeforeFn), dbw.WithAfterWrite(successAfterFn)},
			},
			want:            1,
			wantErr:         false,
			wantName:        "with-before-after-write-success" + id,
			wantPhoneNumber: "updated" + id,
		},
		{
			name: "with-before-write-fail",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "with-before-write-fail" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithBeforeWrite(failedBeforeFn)},
			},
			want:       0,
			wantErr:    true,
			wantErrMsg: "dbw.Update: error before write: fail",
		},
		{
			name: "with-after-write-fail",
			args: args{
				i: &dbtest.TestUser{
					StoreTestUser: &dbtest.StoreTestUser{
						Name:        "with-after-write-fail" + id,
						Email:       "updated" + id,
						PhoneNumber: "updated" + id,
					},
				},
				fieldMaskPaths: []string{"Name", "PhoneNumber"},
				setToNullPaths: []string{"Email"},
				opt:            []dbw.Option{dbw.WithAfterWrite(failedAfterFn)},
			},
			want:       1,
			wantErr:    true,
			wantErrMsg: "dbw.Update: error after write: fail",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			rw := dbw.New(conn)
			u := testUser(t, rw, tt.name+id, id, id)

			if tt.args.i != nil {
				tt.args.i.PublicId = u.PublicId
			}
			rowsUpdated, err := rw.Update(context.Background(), tt.args.i, tt.args.fieldMaskPaths, tt.args.setToNullPaths, tt.args.opt...)
			if tt.wantErr {
				require.Error(err)
				assert.Equal(tt.want, rowsUpdated)
				assert.Contains(err.Error(), tt.wantErrMsg)
				return
			}
			require.NoError(err)
			assert.Equal(tt.want, rowsUpdated)
			if tt.want == 0 {
				return
			}
			foundUser, err := dbtest.NewTestUser()
			require.NoError(err)
			foundUser.PublicId = tt.args.i.PublicId
			where := "public_id = ?"
			for _, f := range tt.args.setToNullPaths {
				switch {
				case strings.EqualFold(f, "phonenumber"):
					f = "phone_number"
				}
				where = fmt.Sprintf("%s and %s is null", where, f)
			}
			err = rw.LookupWhere(context.Background(), foundUser, where, tt.args.i.PublicId)
			require.NoError(err)
			assert.Equal(tt.args.i.PublicId, foundUser.PublicId)
			assert.Equal(tt.wantName, foundUser.Name)
			assert.Equal(tt.wantEmail, foundUser.Email)
			assert.Equal(tt.wantPhoneNumber, foundUser.PhoneNumber)
			assert.NotEqual(now, foundUser.CreateTime)
			assert.NotEqual(now, foundUser.UpdateTime)
			assert.NotEqual(publicId, foundUser.PublicId)
			assert.Equal(u.Version+1, foundUser.Version)
		})
	}
	t.Run("no-version-field", func(t *testing.T) {
		assert := assert.New(t)
		testCarFn := func(t *testing.T, rw *dbw.RW, name, model string, mpg int32) *dbtest.TestCar {
			t.Helper()
			require := require.New(t)
			c, err := dbtest.NewTestCar()
			require.NoError(err)
			c.Name = name
			c.Model = model
			c.Mpg = mpg
			if rw != nil {
				err = rw.Create(context.Background(), c)
				require.NoError(err)
			}
			return c
		}
		w := dbw.New(conn)
		id, err := dbw.NewId("id")
		assert.NoError(err)
		car := testCarFn(t, w, "foo-"+id, id, int32(100))

		car.Name = "friendly-" + id
		versionOne := uint32(1)
		rowsUpdated, err := w.Update(context.Background(), car, []string{"Name"}, nil, dbw.WithVersion(&versionOne))
		assert.Error(err)
		assert.Equal(0, rowsUpdated)
	})
	t.Run("vet-for-write", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		user := testUser(t, w, id, id, id)

		user.Name = "friendly-" + id
		rowsUpdated, err := w.Update(context.Background(), user, []string{"Name"}, nil)
		require.NoError(err)
		assert.Equal(1, rowsUpdated)

		foundUser := dbtest.AllocTestUser()
		foundUser.PublicId = user.PublicId
		err = w.LookupByPublicId(context.Background(), &foundUser)
		require.NoError(err)
		assert.Equal(foundUser.Name, user.Name)

		user.Name = "fail-VetForWrite"
		rowsUpdated, err = w.Update(context.Background(), user, []string{"Name"}, nil)
		require.Error(err)
		assert.Equal(0, rowsUpdated)
		assert.Contains(err.Error(), "fail-VetForWrite")
	})
	t.Run("nil-tx", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := &dbw.RW{}

		user, err := dbtest.NewTestUser()
		require.NoError(err)
		rowsUpdated, err := w.Update(context.Background(), user, []string{"Name"}, nil)
		assert.Error(err)
		assert.Equal(0, rowsUpdated)
		assert.Contains(err.Error(), "dbw.Update: missing underlying db: invalid parameter")
	})
	t.Run("multi-column", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		user := testUser(t, w, "", "", "")
		car := testCar(t, w)
		rental, err := dbtest.NewTestRental(user.PublicId, car.PublicId)
		require.NoError(err)
		require.NoError(w.Create(context.Background(), rental))
		rental.Name = "great rental"
		rowsUpdated, err := w.Update(context.Background(), rental, []string{"Name"}, nil)
		require.NoError(err)
		assert.Equal(1, rowsUpdated)
	})
	t.Run("primary-key", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		user := testUser(t, w, "", "", "")
		car := testCar(t, w)
		rental, err := dbtest.NewTestRental(user.PublicId, car.PublicId)
		require.NoError(err)
		require.NoError(w.Create(context.Background(), rental))
		rental.UserId = "not-allowed"
		rowsUpdated, err := w.Update(context.Background(), rental, []string{"UserId"}, nil)
		require.Error(err)
		assert.Equal(0, rowsUpdated)
		assert.Equal("dbw.Update: not allowed on primary key field UserId: invalid field mask", err.Error())
	})
	t.Run("primary-key-is-zero", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		w := dbw.New(conn)
		user := testUser(t, w, "", "", "")
		car := testCar(t, w)
		rental, err := dbtest.NewTestRental(user.PublicId, car.PublicId)
		require.NoError(err)
		require.NoError(w.Create(context.Background(), rental))
		rental.UserId = ""
		rowsUpdated, err := w.Update(context.Background(), rental, nil, []string{"Name"})
		require.Error(err)
		assert.Equal(0, rowsUpdated)
		assert.Equal("dbw.Update: primary key is not set for: [UserId]: invalid parameter", err.Error())
	})
	t.Run("hooks", func(t *testing.T) {
		hookTests := []struct {
			name     string
			resource interface{}
		}{
			{"before-update", &dbtest.TestWithBeforeUpdate{}},
			{"after-update", &dbtest.TestWithAfterUpdate{}},
		}
		for _, tt := range hookTests {
			t.Run(tt.name, func(t *testing.T) {
				assert, require := assert.New(t), require.New(t)
				w := dbw.New(conn)
				rowsUpdated, err := w.Update(context.Background(), tt.resource, []string{"Name"}, nil)
				require.Error(err)
				assert.ErrorIs(err, dbw.ErrInvalidParameter)
				assert.Contains(err.Error(), "gorm callback/hooks are not supported")
				assert.Equal(0, rowsUpdated)
			})
		}
	})
}
