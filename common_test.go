package dbw_test

import (
	"testing"

	"github.com/hashicorp/go-dbw"
	"github.com/hashicorp/go-dbw/internal/dbtest"
	"github.com/hashicorp/go-secure-stdlib/base62"
	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func TestUpdateFields(t *testing.T) {
	a := assert.New(t)
	id, err := uuid.GenerateUUID()
	a.NoError(err)

	testPublicIdFn := func(t *testing.T) string {
		t.Helper()
		publicId, err := base62.Random(20)
		assert.NoError(t, err)
		return publicId
	}
	testUserFn := func(t *testing.T, name, email string) *dbtest.TestUser {
		t.Helper()
		return &dbtest.TestUser{
			StoreTestUser: &dbtest.StoreTestUser{
				PublicId: testPublicIdFn(t),
				Name:     name,
				Email:    email,
			},
		}
	}

	type args struct {
		i              interface{}
		fieldMaskPaths []string
		setToNullPaths []string
	}
	tests := []struct {
		name       string
		args       args
		want       map[string]interface{}
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "missing interface",
			args: args{
				i:              nil,
				fieldMaskPaths: []string{},
				setToNullPaths: []string{},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: interface is missing: invalid parameter",
		},
		{
			name: "missing fieldmasks",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: nil,
				setToNullPaths: []string{},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: both fieldMaskPaths and setToNullPaths are zero len: invalid parameter",
		},
		{
			name: "missing null fields",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{"Name"},
				setToNullPaths: nil,
			},
			want: map[string]interface{}{
				"Name": id,
			},
			wantErr: false,
		},
		{
			name: "all zero len",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{},
				setToNullPaths: nil,
			},
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: both fieldMaskPaths and setToNullPaths are zero len: invalid parameter",
		},
		{
			name: "not found masks",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{"invalidFieldName"},
				setToNullPaths: []string{},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: field mask paths not found in resource: [invalidFieldName]: invalid parameter",
		},
		{
			name: "not found null paths",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{"name"},
				setToNullPaths: []string{"invalidFieldName"},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: null paths not found in resource: [invalidFieldName]: invalid parameter",
		},
		{
			name: "intersection",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{"name"},
				setToNullPaths: []string{"name"},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: fieldMashPaths and setToNullPaths cannot intersect: invalid parameter",
		},
		{
			name: "valid",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{"name"},
				setToNullPaths: []string{"email"},
			},
			want: map[string]interface{}{
				"name":  id,
				"email": gorm.Expr("NULL"),
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "valid-just-masks",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{"name", "email"},
				setToNullPaths: []string{},
			},
			want: map[string]interface{}{
				"name":  id,
				"email": id,
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "valid-just-nulls",
			args: args{
				i:              testUserFn(t, id, id),
				fieldMaskPaths: []string{},
				setToNullPaths: []string{"name", "email"},
			},
			want: map[string]interface{}{
				"name":  gorm.Expr("NULL"),
				"email": gorm.Expr("NULL"),
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "valid-not-embedded",
			args: args{
				i: dbtest.StoreTestUser{
					PublicId: testPublicIdFn(t),
					Name:     id,
					Email:    "",
				},
				fieldMaskPaths: []string{"name"},
				setToNullPaths: []string{"email"},
			},
			want: map[string]interface{}{
				"name":  id,
				"email": gorm.Expr("NULL"),
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "valid-not-embedded-just-masks",
			args: args{
				i: dbtest.StoreTestUser{
					PublicId: testPublicIdFn(t),
					Name:     id,
					Email:    "",
				},
				fieldMaskPaths: []string{"name"},
				setToNullPaths: nil,
			},
			want: map[string]interface{}{
				"name": id,
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "valid-not-embedded-just-nulls",
			args: args{
				i: dbtest.StoreTestUser{
					PublicId: testPublicIdFn(t),
					Name:     id,
					Email:    "",
				},
				fieldMaskPaths: nil,
				setToNullPaths: []string{"email"},
			},
			want: map[string]interface{}{
				"email": gorm.Expr("NULL"),
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "not found null paths - not embedded",
			args: args{
				i: dbtest.StoreTestUser{
					PublicId: testPublicIdFn(t),
					Name:     id,
					Email:    "",
				},
				fieldMaskPaths: []string{"name"},
				setToNullPaths: []string{"invalidFieldName"},
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "dbw.UpdateFields: null paths not found in resource: [invalidFieldName]: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got, err := dbw.UpdateFields(tt.args.i, tt.args.fieldMaskPaths, tt.args.setToNullPaths)
			if err == nil && tt.wantErr {
				assert.Error(err)
			}
			if tt.wantErr {
				assert.Error(err)
				assert.Equal(tt.wantErrMsg, err.Error())
			}
			assert.Equal(tt.want, got)
		})
	}
	t.Run("valid-embedded-timestamp", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		wantTs := &dbtest.Timestamp{
			Timestamp: &timestamppb.Timestamp{
				Seconds: 1,
				Nanos:   1,
			},
		}
		u := testUserFn(t, "", "")
		u.UpdateTime = wantTs
		got, err := dbw.UpdateFields(u, []string{"UpdateTime"}, nil)
		require.NoError(err)
		assert.True(proto.Equal(wantTs, got["UpdateTime"].(*dbtest.Timestamp)))
	})
}

func TestBuildUpdatePaths(t *testing.T) {
	type args struct {
		fieldValues     map[string]interface{}
		fieldMask       []string
		allowZeroFields []string
	}
	tests := []struct {
		name      string
		args      args
		wantMasks []string
		wantNulls []string
	}{
		{
			name: "empty-inputs",
			args: args{
				fieldValues:     map[string]interface{}{},
				fieldMask:       []string{},
				allowZeroFields: []string{},
			},
			wantMasks: []string{},
			wantNulls: []string{},
		},
		{
			name: "no-changes",
			args: args{
				fieldValues: map[string]interface{}{
					"Boolean":       true,
					"Int":           100,
					"String":        "hello",
					"Float":         1.1,
					"Complex":       complex(1.1, 1.1),
					"ByteSlice":     []byte("byte slice"),
					"ZeroBoolean":   false,
					"ZeroInt":       0,
					"ZeroString":    "",
					"ZeroFloat":     0.0,
					"ZeroComplex":   complex(0.0, 0.0),
					"ZeroByteSlice": nil,
				},
				fieldMask:       []string{},
				allowZeroFields: []string{},
			},
			wantMasks: []string{},
			wantNulls: []string{},
		},
		{
			name: "empty-field-mask-allow-all-zero-fields",
			args: args{
				fieldValues: map[string]interface{}{
					"Boolean":       true,
					"Int":           100,
					"String":        "hello",
					"Float":         1.1,
					"Complex":       complex(1.1, 1.1),
					"ByteSlice":     []byte("byte slice"),
					"ZeroBoolean":   false,
					"ZeroInt":       0,
					"ZeroString":    "",
					"ZeroFloat":     0.0,
					"ZeroComplex":   complex(0.0, 0.0),
					"ZeroByteSlice": nil,
				},
				fieldMask: []string{},
				allowZeroFields: []string{
					"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
					"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
				},
			},
			wantMasks: []string{},
			wantNulls: []string{},
		},
		{
			name: "zero-fields-are-nulls",
			args: args{
				fieldValues: map[string]interface{}{
					"Boolean":       true,
					"Int":           100,
					"String":        "hello",
					"Float":         1.1,
					"Complex":       complex(1.1, 1.1),
					"ByteSlice":     []byte("byte slice"),
					"ZeroBoolean":   false,
					"ZeroInt":       0,
					"ZeroString":    "",
					"ZeroFloat":     0.0,
					"ZeroComplex":   complex(0.0, 0.0),
					"ZeroByteSlice": nil,
				},
				fieldMask: []string{
					"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
					"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
				},
				allowZeroFields: []string{},
			},
			wantMasks: []string{
				"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
			},
			wantNulls: []string{
				"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
			},
		},
		{
			name: "all-zero-fields-allowed-no-nulls",
			args: args{
				fieldValues: map[string]interface{}{
					"Boolean":       true,
					"Int":           100,
					"String":        "hello",
					"Float":         1.1,
					"Complex":       complex(1.1, 1.1),
					"ByteSlice":     []byte("byte slice"),
					"ZeroBoolean":   false,
					"ZeroInt":       0,
					"ZeroString":    "",
					"ZeroFloat":     0.0,
					"ZeroComplex":   complex(0.0, 0.0),
					"ZeroByteSlice": nil,
				},
				fieldMask: []string{
					"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
					"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
				},
				allowZeroFields: []string{
					"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
				},
			},
			wantMasks: []string{
				"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
				"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
			},
			wantNulls: []string{},
		},
		{
			name: "non-zeros-allowed-as-zero-fields",
			args: args{
				fieldValues: map[string]interface{}{
					"Boolean":       true,
					"Int":           100,
					"String":        "hello",
					"Float":         1.1,
					"Complex":       complex(1.1, 1.1),
					"ByteSlice":     []byte("byte slice"),
					"ZeroBoolean":   false,
					"ZeroInt":       0,
					"ZeroString":    "",
					"ZeroFloat":     0.0,
					"ZeroComplex":   complex(0.0, 0.0),
					"ZeroByteSlice": nil,
				},
				fieldMask: []string{
					"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
					"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
				},
				allowZeroFields: []string{
					"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
				},
			},
			wantMasks: []string{
				"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
			},
			wantNulls: []string{
				"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
			},
		},
		{
			name: "only-zero-fields-in-fieldmask",
			args: args{
				fieldValues: map[string]interface{}{
					"Boolean":       true,
					"Int":           100,
					"String":        "hello",
					"Float":         1.1,
					"Complex":       complex(1.1, 1.1),
					"ByteSlice":     []byte("byte slice"),
					"ZeroBoolean":   false,
					"ZeroInt":       0,
					"ZeroString":    "",
					"ZeroFloat":     0.0,
					"ZeroComplex":   complex(0.0, 0.0),
					"ZeroByteSlice": nil,
				},
				fieldMask: []string{
					"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
				},
				allowZeroFields: []string{
					"Boolean", "Int", "String", "Float", "Complex", "ByteSlice",
				},
			},
			wantMasks: []string{},
			wantNulls: []string{
				"ZeroBoolean", "ZeroInt", "ZeroString", "ZeroFloat", "ZeroComplex", "ZeroByteSlice",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			gotMasks, gotNulls := dbw.BuildUpdatePaths(tt.args.fieldValues, tt.args.fieldMask, tt.args.allowZeroFields)
			assert.ElementsMatch(tt.wantMasks, gotMasks, "masks")
			assert.ElementsMatch(tt.wantNulls, gotNulls, "nulls")
		})
	}
}
