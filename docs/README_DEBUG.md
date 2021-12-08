# Debug output
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` provides a few ways to get debug output from the underlying database.

## [DB.Debug(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#DB.Debug)

```go
// enable debug output for all database operations
db, err := dbw.Open(dbw.Sqlite, "dbw.db") 
db.Debug(true)
```

## [WithDebug(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithDebug) 
Operations may take the
[WithDebug(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithDebug)
option which will enable/disable debug output for the duration of that operation. 

```go
// enable debug output for a create operation
rw.Create(ctx, &user, dbw.WithDebug(true))
```