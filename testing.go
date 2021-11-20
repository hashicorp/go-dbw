package db

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/xo/dburl"

	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

// setup the tests (initialize the database one-time). Do not close the returned
// db.  Supported test options: WithDebug, WithTestDialect, WithTestDatabaseUrl,
// and WithMigration
func TestSetup(t *testing.T, opt ...TestOption) (*DB, string) {
	require := require.New(t)
	var url string
	var err error
	ctx := context.Background()

	opts := getTestOpts(opt...)

	switch strings.ToLower(os.Getenv("DB_DIALECT")) {
	case "postgres":
		opts.withDialect = Postgres.String()
	case "sqlite":
		opts.withDialect = Sqlite.String()
	default:
		if opts.withDialect == "" {
			opts.withDialect = Sqlite.String()
		}
	}

	if url := os.Getenv("DB_DSN"); url != "" {
		opts.withTestDatabaseUrl = url
	}

	switch {
	case opts.withDialect == Postgres.String() && opts.withTestDatabaseUrl == "":
		t.Fatal("missing postgres test db url")

	case opts.withDialect == Sqlite.String() && opts.withTestDatabaseUrl == "":
		tmpDbFile, err := ioutil.TempFile("./", "tmp-db")
		require.NoError(err)
		t.Cleanup(func() {
			os.Remove(tmpDbFile.Name())
			os.Remove(tmpDbFile.Name() + "-journal")
		})
		url = tmpDbFile.Name()

	default:
		url = opts.withTestDatabaseUrl
	}

	switch opts.withDialect {
	case Postgres.String():
		u, err := dburl.Parse(opts.withTestDatabaseUrl)
		require.NoError(err)
		db, err := Open(Postgres, u.DSN)
		require.NoError(err)
		rw := New(db)
		tmpDbName, err := newId("go-db-tmp")
		require.NoError(err)
		_, err = rw.Exec(ctx, fmt.Sprintf(`create database %s`, tmpDbName), nil)
		require.NoError(err)
		t.Cleanup(func() {
			_, err = rw.Exec(ctx, fmt.Sprintf(`drop database %s`, tmpDbName), nil)
			assert.NoError(t, err)
		})
		_, err = rw.Exec(ctx, fmt.Sprintf("grant all privileges on database %s to %s", tmpDbName, u.User), nil)
		require.NoError(err)

		namesSegs := strings.Split(strings.TrimPrefix(u.Path, "/"), "?")
		require.Truef(len(namesSegs) > 0, "couldn't determine db name from URL")
		namesSegs[0] = tmpDbName
		u.Path = strings.Join(namesSegs, "?")
		opts.withTestDatabaseUrl, err = dburl.GenPostgres(u)
		require.NoError(err)
	}

	dbType, err := StringToDbType(opts.withDialect)
	require.NoError(err)

	db, err := Open(dbType, url)
	require.NoError(err)

	db.Logger.LogMode(logger.Error)
	t.Cleanup(func() {
		sqlDB, err := db.SqlDB(ctx)
		assert.NoError(t, err)
		assert.NoError(t, sqlDB.Close(), "Got error closing db.")
	})

	if opts.withTestDebug || strings.ToLower(os.Getenv("DEBUG")) == "true" {
		db.Debug(true)
	}

	if opts.withTestMigration != nil {
		err = opts.withTestMigration(ctx, opts.withDialect, url)
		require.NoError(err)
	}
	TestCreateTables(t, db)
	return db, url
}

// getTestOpts - iterate the inbound TestOptions and return a struct
func getTestOpts(opt ...TestOption) testOptions {
	opts := getDefaultTestOptions()
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

// TestOption - how Options are passed as arguments
type TestOption func(*testOptions)

// options = how options are represented
type testOptions struct {
	withDialect         string
	withTestDatabaseUrl string
	withTestMigration   func(ctx context.Context, dialect, url string) error
	withTestDebug       bool
}

func getDefaultTestOptions() testOptions {
	return testOptions{}
}

// WithTestDialect provides a way to specify the test database dialect
func WithTestDialect(dialect string) TestOption {
	return func(o *testOptions) {
		o.withDialect = dialect
	}
}

// WithTestMigrationProvides a way to specify an option func which runs a
// required database migration to initialize the database
func WithTestMigration(migrationFn func(ctx context.Context, dialect, url string) error) TestOption {
	return func(o *testOptions) {
		o.withTestMigration = migrationFn
	}
}

// WithTestDatabaseUrl provides a way to specify an existing database for tests
func WithTestDatabaseUrl(url string) TestOption {
	return func(o *testOptions) {
		o.withTestDatabaseUrl = url
	}
}

func TestId(t *testing.T) string {
	t.Helper()
	require := require.New(t)
	id, err := uuid.GenerateUUID()
	require.NoError(err)
	return id
}

// TestCreateTables will create the test tables for the db pkg
func TestCreateTables(t *testing.T, conn *DB) {
	t.Helper()
	t.Cleanup(func() { testDropTables(t, conn) })
	require := require.New(t)
	testCtx := context.Background()
	rw := New(conn)
	switch conn.Dialector.Name() {
	case "sqlite":
		_, err := rw.Exec(testCtx, testQueryCreateTablesSqlite, nil)
		require.NoError(err)
	default:
		t.Fatalf("unknown dialect: %s", conn.Dialector.Name())
	}
}

func testDropTables(t *testing.T, conn *DB) {
	t.Helper()
	require := require.New(t)
	testCtx := context.Background()
	rw := New(conn)
	switch conn.Dialector.Name() {
	case "sqlite":
		_, err := rw.Exec(testCtx, testQueryDropTablesSqlite, nil)
		require.NoError(err)
	default:
		t.Fatalf("unknown dialect: %s", conn.Dialector.Name())
	}
}

const (
	testQueryCreateTablesSqlite = `	
begin;

-- create test tables used in the unit tests for the db package 
-- these tables (db_test_user, db_test_car, db_test_rental, db_test_scooter) are
-- not part of the application's domain model... they are simply used for testing
-- the db package 
create table if not exists db_test_user (
  public_id text primary key,
  create_time wt_timestamp,
  update_time wt_timestamp,
  name text unique,
  phone_number text,
  email text,
  version wt_version
);

create trigger update_time_column 
before update on db_test_user 
for each row 
when 
  new.public_id 	<> old.public_id or
  new.name      	<> old.name or
  new.phone_number 	<> old.phone_number or
  new.email     	<> old.email or
  new.version   	<> old.version 
  begin
    update db_test_user set update_time = datetime('now','localtime') where rowid == new.rowid;
  end;

create trigger immutable_columns 
before update on db_test_user 
for each row 
  when 
	new.create_time <> old.create_time
	begin
	  select raise(abort, 'immutable column');
	end;

create trigger default_create_time_column
before insert on db_test_user
for each row
  begin
	update db_test_user set create_time = datetime('now','localtime') where rowid = new.rowid;
  end;
	
create trigger update_version_column
after update on db_test_user
for each row
when 
  new.public_id 	<> old.public_id or
  new.name      	<> old.name or
  new.phone_number  <> old.phone_number or
  new.email     	<> old.email
  begin
    update db_test_user set version = old.version + 1 where rowid = new.rowid;
  end;
  
  commit;
	`
	testQueryDropTablesSqlite = `
begin;
drop table if exists db_test_user;
commit;
`
)
