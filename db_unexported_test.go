// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dbw

import (
	"bytes"
	"errors"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func TestDB_Debug(t *testing.T) {
	tests := []struct {
		name   string
		enable bool
	}{
		{name: "enabled", enable: true},
		{name: "disabled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			db, _ := TestSetup(t)
			db.Debug(tt.enable)
			if tt.enable {
				assert.Equal(db.wrapped.Logger, logger.Default.LogMode(logger.Info))
			} else {
				assert.Equal(db.wrapped.Logger, logger.Default.LogMode(logger.Error))
			}
		})
	}
}

func TestDB_gormLogger(t *testing.T) {
	var buf bytes.Buffer
	l := getGormLogger(
		hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Trace,
			Output: &buf,
		}),
	)
	t.Run("no-output", func(t *testing.T) {
		l.Printf("not a pgerror", "value 0 placeholder", errors.New("test"), "values 2 placeholder")
		assert.Empty(t, buf.Bytes())
	})
	t.Run("output", func(t *testing.T) {
		l.Printf("is a pgerror", "value 0 placeholder", &pgconn.PgError{}, "values 2 placeholder")
		assert.NotEmpty(t, buf.Bytes())
	})
}
