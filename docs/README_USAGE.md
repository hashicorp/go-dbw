# Usage highlights
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

Just some high-level usage highlights to get you started. Read the [dbw package
docs](https://pkg.go.dev/github.com/hashicorp/go-dbw) for 
a complete list of capabilities and their documentation.

```go
// initialize fields which cannot be set during creation
dbw.InitNonCreatableFields([]string{"CreateTime", "UpdateTime"})

// initialize fields which cannot be updated
dbw.InitNonUpdatableFields([]string{"PublicId", "CreateTime", "UpdateTime"})

// errors are intentionally ignored for brevity 
db, _ := dbw.Open(dialect, url)    
rw := dbw.New(conn)

id, _ := dbw.NewId("u")
user, _ := dbtest.NewTestUser()
_ = rw.Create(context.Background(), user)

foundUser, _ := dbtest.NewTestUser()
foundUser.PublicId = id
_ = rw.LookupBy(context.Background(), foundUser)

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
    []any{ sql.Named{"names", "alice", "bob"}},
)
defer rows.Close()
for rows.Next() {
    user := db_test.NewTestUser()
    _ = rw.ScanRows(rows, &user)
    // Do something with the user struct
}

user.Name = "Alice"
retryErrFn := func(_ error) bool { return true }
_, err = w.DoTx(
    context.Background(),
    func(_ error) bool { return true }, // retry all errors
    3,                                  // three retries
    ExpBackoff{},                       // exponential backoff
    func(w Writer) error {
        // the TxHandler updates the user's name
        _, err := w.Update(context.Background(), 
            user, 
            []string{"Name"}, 
            nil,
            dbw.WithVersion(&user.Version),
        )
        if err != nil {
            return err
        }
    },
)
```
