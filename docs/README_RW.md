# Readers and Writers
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

[RW](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW) provides a type which
implements both the interfaces of
[dbw.Reader](https://pkg.go.dev/github.com/hashicorp/go-dbw#Reader) and
[dbw.Writer](https://pkg.go.dev/github.com/hashicorp/go-dbw#Writer). Many
[RWs](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW)  
can (and likely should) share the same
[dbw.DB](https://pkg.go.dev/github.com/hashicorp/go-dbw#DB), since the
[dbw.DB](https://pkg.go.dev/github.com/hashicorp/go-dbw#DB)
is responsible for connection pooling.   

```go
db, _ := dbw.Open(dbw.Sqlite, url)    
rw := dbw.New(conn)
// now you can use the rw for read/write database operations
```

When required, you can create two `DB`s: one for reading from read replicas and
another for writing to the primary database.  In such a scenario, you'd need to
create RWs with the correct DB for either reading or writing. 

```go
readReplicaDSN := "postgresql://go_db:go_db@reader.hostname:9920/go_db?sslmode=disable"
rdb, err := dbw.Open(dbw.Postgres, readReplicaDSN)    
reader := dbw.New(rdb)


primaryDSN := "postgresql://go_db:go_db@primary.hostname:9920/go_db?sslmode=disable"
rdb, err := dbw.Open(dbw.Postgres, primaryDSN)    
writer := dbw.New(rdb)
```
