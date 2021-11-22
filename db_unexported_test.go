package dbw

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert, require := assert.New(t), require.New(t)
			tmpDbFile, err := ioutil.TempFile("./", "tmp-db")
			require.NoError(err)
			t.Cleanup(func() {
				os.Remove(tmpDbFile.Name())
				os.Remove(tmpDbFile.Name() + "-journal")
			})
			db, err := Open(Sqlite, tmpDbFile.Name())
			require.NoError(err)
			db.Debug(tt.enable)
			if tt.enable {
				assert.Equal(db.Logger, logger.Default.LogMode(logger.Info))
			} else {
				assert.Equal(db.Logger, logger.Default.LogMode(logger.Error))
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
