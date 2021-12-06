# Queries
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` provides quite a few different ways to read resources from a database
(see: [Read operations](./README_READ.md))  

`dbw` intentionally doesn't support "associations" or try to reinvent sql by providing some sort of pattern for
"building" a query.  Instead, `dbw` provides a set of functions for directly issuing SQL to the database and scanning the results back into Go structs. 


## Query example with CTE
```go
where := `
with user_rentals as (
    select user_id, count(*) as rental_count
    from test_rentals
    group by user_id)
select u.public_id, u.name, r.rental_count 
    from test_users u 
join user_rental r
    on u.public_id = r.user_id 
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
```

## Exec example

```go
where := `
delete from test_rentals 
where user_id not in (select user_id from test_users)`

err := rw.Exec(
    context.Background(), 
    where, 
    nil,
)
```