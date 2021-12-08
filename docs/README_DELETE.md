# Delete
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)
## [RW.Delete(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW.Delete) example with one item
```go
err, rowsAffected = rw.Delete(ctx, 
    &user, 
    dbw.WithVersion(&user.Version),
)  
```
## [RW.DeleteItems(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW.DeleteItems) example with multiple items
```go
var rowsAffected int64
err = rw.DeleteItems(ctx,
    []interface{}{&user1, &user2}, 
    dbw.WithRowsAffected(&rowsAffected),
)  
```
