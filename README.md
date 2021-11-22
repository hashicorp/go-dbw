# dbw package

## Usage
Just some high-level usage highlights to get you started.  Read the godocs for 
a complete list of capabilities and their documentation.

```go
    conn, _ := dbw.Open(dialect, &gorm.Config{})    
    rw := Db{Tx: conn}
    
    // There are writer methods like: Create, Update and Delete
    // that will write a struct with Gorm tags to the db.  The caller is 
    // responsible for the transaction life cycle of the writer    
    // and if an error is returned the caller must decide what to do with 
    // the transaction, which is almost always a rollback for the caller.
    err = rw.Create(context.Background(), user)
   
    // There are reader methods like: LookupByPublicId,  
    // LookupByName, SearchWhere, LookupWhere, etc
    // which will lookup resources for you and scan them into your struct
    // with Gorm tags
    err = rw.LookupByPublicId(context.Background(), foundUser)

    // There's reader ScanRows that facilitates scanning rows from 
    // a query into your struct with Gorm tags
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
    rows, err := rw.Query(context.Background(), where, []interface{}{ sql.Named{"names", "alice", "bob"}})
	defer rows.Close()
	for rows.Next() {
        user := db_test.NewTestUser()
        // scan the row into your Gorm struct
		if err := rw.ScanRows(rows, &user); err != nil {
            return err
        }
        // Do something with the user struct
    }

    // DoTx is a writer function that wraps a TxHandler 
    // in a retryable transaction.  You simply implement a
    // TxHandler that does your writes and hand the handler
    // to DoTx, which will wrap the writes in a retryable 
    // transaction with the retry attempts and backoff
    // strategy you specify via options.
    _, err = w.DoTx(
        context.Background(),
        10,           // ten retries
        ExpBackoff{}, // exponential backoff
        func(w Writer) error {
            // the TxHandler updates the user's friendly name
            return w.Update(context.Background(), user, []string{"FriendlyName"})
        })
```
