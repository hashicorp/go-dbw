package dbw

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRW_whereClausesFromOpts(t *testing.T) {
	db, _ := TestSetup(t)
	testCtx := context.Background()
	type testUser struct {
		Version int
	}

	tests := []struct {
		name      string
		rw        *RW
		i         interface{}
		opts      Options
		wantWhere string
		wantArgs  []interface{}
		wantErr   bool
	}{
		{
			name: "with-version-with-table-on-conflict",
			rw:   New(db),
			i:    &testUser{},
			opts: Options{
				WithVersion:    func() *uint32 { i := uint32(1); return &i }(),
				WithTable:      "test_table",
				WithOnConflict: &OnConflict{},
			},
			wantWhere: "test_table.version = ?",
			wantArgs:  []interface{}{func() *uint32 { i := uint32(1); return &i }()},
		},
		{
			name: "with-version-with-table",
			rw:   New(db),
			i:    &testUser{},
			opts: Options{
				WithVersion: func() *uint32 { i := uint32(1); return &i }(),
				WithTable:   "test_table",
			},
			wantWhere: "version = ?",
			wantArgs:  []interface{}{func() *uint32 { i := uint32(1); return &i }()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			where, whereArgs, err := tt.rw.whereClausesFromOpts(testCtx, tt.i, tt.opts)
			if tt.wantErr {
				require.NoError(err)
				assert.Empty(where)
				assert.Empty(whereArgs)
				return
			}
			require.NoError(err)
			assert.Equal(tt.wantWhere, where)
			assert.Equal(tt.wantArgs, whereArgs)
		})
	}
}
