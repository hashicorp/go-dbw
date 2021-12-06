# Delete
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

## Example with one item
```go
err, rowsAffected = rw.Delete(ctx, 
    &user, 
    dbw.WithVersion(&user.Version),
)  
```
## Example with multiple items
```go
var rowsAffected int64
err = rw.DeleteItems(ctx,
    []interface{}{&user1, &user2}, 
    dbw.WithRowsAffected(&rowsAffected),
)  
```
