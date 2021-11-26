# dbw package

dbw is a database wrapper that supports connecting and using any database with a
[GORM](https://github.com/go-gorm/gorm) driver.  It's intent is to completely
encapsulate an application's access to it's database with the exception of
migrations.    

dbw is intentionally not an ORM and it removes typical ORM abstractions like
"advanced query building", associations and migrations.  

This is not to say you can't easily use dbw for complicated queries, it's just
that dbw doesn't try to reinvent sql by providing some sort of pattern for
building them with functions. Of course, dbw also provides lookup/search
functions when you simply need to read resources from the database.

dbw strives to make CRUD for database resources fairly trivial.  Even supporting
"on conflict" for its create function.  dbw also allows you to opt out of its
CRUD functions and use exec, query and scan rows directly.  You may want to
carefully weigh when it's appropriate to use exec and query directly, since
it's likely that each time you use them you're leaking a bit of your
database schema into your application's domain. 

## Usage
Just some high-level usage highlights to get you started.  Read the godocs for 
a complete list of capabilities and their documentation.

```go
    // errors are intentionally ignored for brevity 
    conn, _ := dbw.Open(dialect, url)    
    rw := dbw.New(conn)

    id, _ := dbw.NewId("u")
    user, _ := dbtest.NewTestUser()
    _ = rw.Create(context.Background(), user)
   
   foundUser, _ := dbtest.NewTestUser()
   foundUser.PublicId = id
    _ = rw.LookupByPublicId(context.Background(), foundUser)

    where := `
    with avg_version as (
        select public_id, avg(version) as avg_version_for_user
        from test_users
        group by version)
    select u.public_id, u.name, av.avg_version_for_user 
    from test_users u 
    join avg_version av
      on u.public_id = av.public_id 
    where name in (@names)`
    rows, err := rw.Query(
        context.Background(), 
        where, 
        []interface{}{ sql.Named{"names", "alice", "bob"}},
    )
	defer rows.Close()
	for rows.Next() {
        user := db_test.NewTestUser()
		_ = rw.ScanRows(rows, &user)
        // Do something with the user struct
    }

    retryErrFn := func(_ error) bool { return true }
    _, err = w.DoTx(
        context.Background(),
        func(_ error) bool { return true }, // retry all errors
        3,                                  // three retries
        ExpBackoff{},                       // exponential backoff
        func(w Writer) error {
            // the TxHandler updates the user's name
            return w.Update(context.Background(), user, []string{"Name"})
        })
```
