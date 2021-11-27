# Debug output
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` provides a few ways to get debug output from the underlying database.

## Enable/Disable debug output for all operations.
```go
// enable debug output for all database operations
db, err := dbw.Open(dbw.Sqlite, "dbw.db") 
db.Debug(true)
```

## WithDebug(...) option
Operations make take the WithDebug(...) option which will enable/disable debug
output for the operation. 

```go
// enable debug output for a create operation
rw.Create(ctx, &user, dbw.WithDebug(true))
```