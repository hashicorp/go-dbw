package dbw

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-secure-stdlib/base62"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getTestOpts(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	t.Run("WithTestMigration", func(t *testing.T) {
		fn := func(context.Context, string, string) error { return nil }
		opts := getTestOpts(WithTestMigration(fn))
		testOpts := getDefaultTestOptions()
		testOpts.withTestMigration = fn
		assert.NotNil(opts, testOpts.withTestMigration)
	})
	t.Run("WithTestDatabaseUrl", func(t *testing.T) {
		opts := getTestOpts(WithTestDatabaseUrl("url"))
		testOpts := getDefaultTestOptions()
		testOpts.withTestDatabaseUrl = "url"
		assert.Equal(opts, testOpts)
	})
}

func Test_TestSetup(t *testing.T) {
	testMigrationFn := func(context.Context, string, string) error {
		conn, err := Open(Sqlite, "file::memory:")
		require.NoError(t, err)
		rw := New(conn)
		_, err = rw.Exec(context.Background(), testQueryCreateTablesSqlite, nil)
		require.NoError(t, err)
		return nil
	}

	testMigrationUsingDbFn := func(_ context.Context, db *sql.DB) error {
		var sql string
		switch strings.ToLower(os.Getenv("DB_DIALECT")) {
		case "postgres":
			sql = testQueryCreateTablesPostgres
		default:
			sql = testQueryCreateTablesSqlite
		}
		_, err := db.Exec(sql)
		require.NoError(t, err)
		return nil
	}

	tests := []struct {
		name     string
		opt      []TestOption
		validate func(db *DB) bool
	}{
		{
			name: "sqlite-with-migration",
			opt:  []TestOption{WithTestDialect(Sqlite.String()), WithTestMigration(testMigrationFn)},
			// we can't validate this, since WithTestMigration will open a new
			// sqlite connection which will result in a new in-memory db which
			// will only existing during the testMigrationFn... sort of silly,
			// but it does test that the fn is called properly at least.
		},
		{
			name: "sqlite-with-migration-using-db",
			opt:  []TestOption{WithTestDialect(Sqlite.String()), WithTestMigrationUsingDB(testMigrationUsingDbFn)},
			validate: func(db *DB) bool {
				rw := New(db)
				publicId, err := base62.Random(20)
				require.NoError(t, err)
				user := &testUser{
					PublicId: publicId,
				}
				require.NoError(t, err)
				user.Name = "foo-" + user.PublicId
				err = rw.Create(context.Background(), user)
				require.NoError(t, err)
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			db, url := TestSetup(t, tt.opt...)
			if tt.validate != nil {
				assert.True(tt.validate(db))
			}
			assert.NotNil(db)
			assert.NotEmpty(url)
		})
	}
}

func Test_CreateDropTestTables(t *testing.T) {
	t.Run("execute", func(t *testing.T) {
		db, _ := TestSetup(t, WithTestDialect(Sqlite.String()))
		testDropTables(t, db)
		TestCreateTables(t, db)
	})
}

// testUser is require since we can't import dbtest as it creates a circular dep
type testUser struct {
	PublicId    string `gorm:"primaryKey;default:null"`
	Name        string `gorm:"default:null"`
	PhoneNumber string `gorm:"default:null"`
	Email       string `gorm:"default:null"`
	Version     uint32 `gorm:"default:null"`
}

func (u *testUser) TableName() string { return "db_test_user" }
