// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dbw

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

// Test_getOpts provides unit tests for GetOpts and all the options
func Test_getOpts(t *testing.T) {
	t.Parallel()
	t.Run("WithLookup", func(t *testing.T) {
		assert := assert.New(t)
		// test default of true
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithLookup = false
		assert.Equal(opts, testOpts)

		// try setting to false
		opts = GetOpts(WithLookup(true))
		testOpts = getDefaultOptions()
		testOpts.WithLookup = true
		assert.Equal(opts, testOpts)
	})
	t.Run("WithFieldMaskPaths", func(t *testing.T) {
		assert := assert.New(t)
		// test default of []string{}
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithFieldMaskPaths = []string{}
		assert.Equal(opts, testOpts)

		testPaths := []string{"alice", "bob"}
		opts = GetOpts(WithFieldMaskPaths(testPaths))
		testOpts = getDefaultOptions()
		testOpts.WithFieldMaskPaths = testPaths
		assert.Equal(opts, testOpts)
	})
	t.Run("WithNullPaths", func(t *testing.T) {
		assert := assert.New(t)
		// test default of []string{}
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithNullPaths = []string{}
		assert.Equal(opts, testOpts)

		testPaths := []string{"alice", "bob"}
		opts = GetOpts(WithNullPaths(testPaths))
		testOpts = getDefaultOptions()
		testOpts.WithNullPaths = testPaths
		assert.Equal(opts, testOpts)
	})
	t.Run("WithLimit", func(t *testing.T) {
		assert := assert.New(t)
		// test default of 0
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithLimit = 0
		assert.Equal(opts, testOpts)

		opts = GetOpts(WithLimit(-1))
		testOpts = getDefaultOptions()
		testOpts.WithLimit = -1
		assert.Equal(opts, testOpts)

		opts = GetOpts(WithLimit(1))
		testOpts = getDefaultOptions()
		testOpts.WithLimit = 1
		assert.Equal(opts, testOpts)
	})
	t.Run("WithVersion", func(t *testing.T) {
		assert := assert.New(t)
		// test default of 0
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithVersion = nil
		assert.Equal(opts, testOpts)
		versionTwo := uint32(2)
		opts = GetOpts(WithVersion(&versionTwo))
		testOpts = getDefaultOptions()
		testOpts.WithVersion = &versionTwo
		assert.Equal(opts, testOpts)
	})
	t.Run("WithSkipVetForWrite", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithSkipVetForWrite = false
		assert.Equal(opts, testOpts)
		opts = GetOpts(WithSkipVetForWrite(true))
		testOpts = getDefaultOptions()
		testOpts.WithSkipVetForWrite = true
		assert.Equal(opts, testOpts)
	})
	t.Run("WithWhere", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithWhereClause = ""
		testOpts.WithWhereClauseArgs = nil
		assert.Equal(opts, testOpts)
		opts = GetOpts(WithWhere("id = ? and foo = ?", 1234, "bar"))
		testOpts.WithWhereClause = "id = ? and foo = ?"
		testOpts.WithWhereClauseArgs = []interface{}{1234, "bar"}
		assert.Equal(opts, testOpts)
	})
	t.Run("WithOrder", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithOrder = ""
		assert.Equal(opts, testOpts)
		opts = GetOpts(WithOrder("version desc"))
		testOpts.WithOrder = "version desc"
		assert.Equal(opts, testOpts)
	})
	t.Run("WithGormFormatter", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		assert.Equal(opts, testOpts)

		testLogger := hclog.New(&hclog.LoggerOptions{})
		opts = GetOpts(WithLogger(testLogger))
		testOpts.WithLogger = testLogger
		assert.Equal(opts, testOpts)
	})
	t.Run("WithMaxOpenConnections", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		assert.Equal(opts, testOpts)
		opts = GetOpts(WithMaxOpenConnections(22))
		testOpts.WithMaxOpenConnections = 22
		assert.Equal(opts, testOpts)
	})
	t.Run("WithDebug", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		assert.Equal(opts, testOpts)
		// try setting to true
		opts = GetOpts(WithDebug(true))
		testOpts.WithDebug = true
		assert.Equal(opts, testOpts)
	})
	t.Run("WithOnConflict", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		assert.Equal(opts, testOpts)
		columns := SetColumns([]string{"name", "description"})
		columnValues := SetColumnValues(map[string]interface{}{"expiration": "NULL"})
		testOnConflict := OnConflict{
			Target: Constraint("uniq-name"),
			Action: append(columns, columnValues...),
		}
		opts = GetOpts(WithOnConflict(&testOnConflict))
		testOpts.WithOnConflict = &testOnConflict
		assert.Equal(opts, testOpts)
	})
	t.Run("WithReturnRowsAffected", func(t *testing.T) {
		assert := assert.New(t)
		// test default of false
		opts := GetOpts()
		testOpts := getDefaultOptions()
		assert.Equal(opts, testOpts)

		var rowsAffected int64
		opts = GetOpts(WithReturnRowsAffected(&rowsAffected))
		testOpts.WithRowsAffected = &rowsAffected
		assert.Equal(opts, testOpts)
	})
	t.Run("WithBeforeWrite", func(t *testing.T) {
		assert := assert.New(t)
		// test defaults
		opts := GetOpts()
		assert.Nil(opts.WithBeforeWrite)

		fn := func(interface{}) error { return nil }
		opts = GetOpts(WithBeforeWrite(fn))
		assert.NotNil(opts.WithBeforeWrite)
	})
	t.Run("WithAfterWrite", func(t *testing.T) {
		assert := assert.New(t)
		// test defaults
		opts := GetOpts()
		assert.Nil(opts.WithAfterWrite)

		fn := func(interface{}, int) error { return nil }
		opts = GetOpts(WithAfterWrite(fn))
		assert.NotNil(opts.WithAfterWrite)
	})
	t.Run("WithMaxOpenConnections", func(t *testing.T) {
		assert := assert.New(t)
		// test default of 0
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithMaxOpenConnections = 0
		assert.Equal(opts, testOpts)
		opts = GetOpts(WithMaxOpenConnections(1))
		testOpts = getDefaultOptions()
		testOpts.WithMaxOpenConnections = 1
		assert.Equal(opts, testOpts)
	})
	t.Run("WithMinOpenConnections", func(t *testing.T) {
		assert := assert.New(t)
		// test default of 0
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithMinOpenConnections = 0
		assert.Equal(opts, testOpts)
		opts = GetOpts(WithMinOpenConnections(1))
		testOpts = getDefaultOptions()
		testOpts.WithMinOpenConnections = 1
		assert.Equal(opts, testOpts)
	})
	t.Run("WithTable", func(t *testing.T) {
		assert := assert.New(t)
		// test default
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithTable = ""
		assert.Equal(opts, testOpts)

		opts = GetOpts(WithTable("tmp_table_name"))
		testOpts = getDefaultOptions()
		testOpts.WithTable = "tmp_table_name"
		assert.Equal(opts, testOpts)
	})
	t.Run("WithLogLevel", func(t *testing.T) {
		assert := assert.New(t)
		// test default
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.withLogLevel = Error
		assert.Equal(opts, testOpts)

		opts = GetOpts(WithLogLevel(Warn))
		testOpts = getDefaultOptions()
		testOpts.withLogLevel = Warn
		assert.Equal(opts, testOpts)
	})
	t.Run("WithBatchSize", func(t *testing.T) {
		assert := assert.New(t)
		// test default
		opts := GetOpts()
		testOpts := getDefaultOptions()
		testOpts.WithBatchSize = DefaultBatchSize
		assert.Equal(opts, testOpts)

		opts = GetOpts(WithBatchSize(100))
		testOpts = getDefaultOptions()
		testOpts.WithBatchSize = 100
		assert.Equal(opts, testOpts)
	})
}
