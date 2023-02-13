// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func Test_validateResourcesInterface(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		resources       interface{}
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "not-ptr-to-slice",
			resources:       []*string{},
			wantErrContains: "interface parameter must to be a pointer:",
		},
		{
			name:            "not-ptr",
			resources:       "string",
			wantErrContains: "interface parameter must to be a pointer:",
		},
		{
			name:            "not-slice-of-ptrs",
			resources:       &[]string{},
			wantErrContains: "interface parameter is a slice, but the elements of the slice are not pointers",
		},
		{
			name:      "success-ptr-to-slice-of-ptrs",
			resources: &[]*string{},
		},
		{
			name: "success-ptr",
			resources: func() interface{} {
				s := "s"
				return &s
			}(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			err := validateResourcesInterface(tc.resources)
			if tc.wantErr {
				assert.Error(err)
				if tc.wantErrContains != "" {
					assert.Contains(err.Error(), tc.wantErrContains)
				}
			}
		})
	}
}
