package dbw

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-secure-stdlib/base62"
	"github.com/xo/dburl"

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
		tmpDbName, err := newId("go_db_tmp")
		tmpDbName = strings.ToLower(tmpDbName)
		require.NoError(err)
		_, err = rw.Exec(ctx, fmt.Sprintf(`create database "%s"`, tmpDbName), nil)
		require.NoError(err)
		t.Cleanup(func() {
			_, err = rw.Exec(ctx, fmt.Sprintf(`drop database %s`, tmpDbName), nil)
			assert.NoError(t, err)
		})
		_, err = rw.Exec(ctx, fmt.Sprintf(`grant all privileges on database %s to %s`, tmpDbName, u.User.Username()), nil)
		require.NoError(err)

		namesSegs := strings.Split(strings.TrimPrefix(u.Path, "/"), "?")
		require.Truef(len(namesSegs) > 0, "couldn't determine db name from URL")
		namesSegs[0] = tmpDbName
		u.Path = strings.Join(namesSegs, "?")
		url, err = dburl.GenPostgres(u)
		require.NoError(err)
	}

	dbType, err := StringToDbType(opts.withDialect)
	require.NoError(err)

	db, err := Open(dbType, url)
	require.NoError(err)

	db.Logger.LogMode(logger.Error)
	t.Cleanup(func() {
		assert.NoError(t, db.Close(ctx), "Got error closing db.")
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
	id, err := base62.Random(20)
	require.NoError(err)
	return id
}

// TestCreateTables will create the test tables for the db pkg
func TestCreateTables(t *testing.T, conn *DB) {
	t.Helper()
	require := require.New(t)
	testCtx := context.Background()
	rw := New(conn)
	var query string
	switch conn.Dialector.Name() {
	case "sqlite":
		query = testQueryCreateTablesSqlite
	case "postgres":
		query = testQueryCreateTablesPostgres
	default:
		t.Fatalf("unknown dialect: %s", conn.Dialector.Name())
	}
	_, err := rw.Exec(testCtx, query, nil)
	require.NoError(err)
}

func testDropTables(t *testing.T, conn *DB) {
	t.Helper()
	require := require.New(t)
	testCtx := context.Background()
	rw := New(conn)
	var query string
	switch conn.Dialector.Name() {
	case "sqlite":
		query = testQueryDropTablesSqlite
	case "postgres":
		query = testQueryDropTablesPostgres
	default:
		t.Fatalf("unknown dialect: %s", conn.Dialector.Name())
	}
	_, err := rw.Exec(testCtx, query, nil)
	require.NoError(err)
}

const (
	testQueryCreateTablesSqlite = `	
begin;

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

	testQueryCreateTablesPostgres = `
begin;

create domain wt_public_id as text
check(
  length(trim(value)) > 10
);
comment on domain wt_public_id is
'Random ID generated with github.com/hashicorp/go-secure-stdlib/base62';

drop domain if exists wt_timestamp;
create domain wt_timestamp as
  timestamp with time zone
  default current_timestamp;
comment on domain wt_timestamp is
'Standard timestamp for all create_time and update_time columns';

create or replace function
  update_time_column()
  returns trigger
as $$
begin
  if row(new.*) is distinct from row(old.*) then
    new.update_time = now();
    return new;
  else
    return old;
  end if;
end;
$$ language plpgsql;

comment on function
  update_time_column()
is
  'function used in before update triggers to properly set update_time columns';
  
create or replace function
  default_create_time()
  returns trigger
as $$
begin
  if new.create_time is distinct from now() then
    raise warning 'create_time cannot be set to %', new.create_time;
    new.create_time = now();
  end if;
  return new;
end;
$$ language plpgsql;

comment on function
  default_create_time()
is
  'function used in before insert triggers to set create_time column to now';


  create domain wt_version as bigint
  default 1
  not null
  check(
   value > 0
  );
comment on domain wt_version is
'standard column for row version';

-- update_version_column() will increment the version column whenever row data
-- is updated and should only be used in an update after trigger.  This function
-- will overwrite any explicit updates to the version column. The function
-- accepts an optional parameter of 'private_id' for the tables primary key.
create or replace function
  update_version_column()
  returns trigger
as $$
begin
  if pg_trigger_depth() = 1 then
    if row(new.*) is distinct from row(old.*) then
      if tg_nargs = 0 then
        execute format('update %I set version = $1 where public_id = $2', tg_relid::regclass) using old.version+1, new.public_id;
        new.version = old.version + 1;
        return new;
      end if;
      if tg_argv[0] = 'private_id' then
        execute format('update %I set version = $1 where private_id = $2', tg_relid::regclass) using old.version+1, new.private_id;
        new.version = old.version + 1;
        return new;
      end if;
    end if;
  end if;
  return new;
end;
$$ language plpgsql;

comment on function
  update_version_column()
is
  'function used in after update triggers to properly set version columns';

-- immutable_columns() will make the column names immutable which are passed as
-- parameters when the trigger is created. It raises error code 23601 which is a
-- class 23 integrity constraint violation: immutable column  
create or replace function
  immutable_columns()
  returns trigger
as $$
declare 
	col_name text; 
	new_value text;
	old_value text;
begin
  foreach col_name in array tg_argv loop
    execute format('SELECT $1.%I', col_name) into new_value using new;
    execute format('SELECT $1.%I', col_name) into old_value using old;
  	if new_value is distinct from old_value then
      raise exception 'immutable column: %.%', tg_table_name, col_name using
        errcode = '23601', 
        schema = tg_table_schema,
        table = tg_table_name,
        column = col_name;
  	end if;
  end loop;
  return new;
end;
$$ language plpgsql;

comment on function
  immutable_columns()
is
  'function used in before update triggers to make columns immutable';

-- ########################################################################################

create table if not exists db_test_user (
	public_id text primary key,
	create_time wt_timestamp,
	update_time wt_timestamp,
	name text unique,
	phone_number text,
	email text,
	version wt_version
);
	
create trigger 
	update_time_column 
before 
update on db_test_user 
	for each row execute procedure update_time_column();

-- define the immutable fields for db_test_user
create trigger 
	immutable_columns
before
update on db_test_user
	for each row execute procedure immutable_columns('create_time');

create trigger 
	default_create_time_column
before
insert on db_test_user 
	for each row execute procedure default_create_time();

create trigger 
	update_version_column
after update on db_test_user
	for each row execute procedure update_version_column();


commit;
	`

	testQueryDropTablesSqlite = `
begin;
drop table if exists db_test_user;
commit;
`

	testQueryDropTablesPostgres = `
begin;
drop table if exists db_test_user cascade;
drop domain if exists wt_public_id;
drop domain if exists wt_timestamp;
drop domain if exists wt_version;
commit;
	`
)
