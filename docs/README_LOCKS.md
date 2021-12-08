# Optimistic locking for write operations
[![Go Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` provides the [dbw.WithVersion(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithVersion) option for write operations to enable
an optimistic locking pattern.  Using this pattern, the caller must first read
a resource from the database and get its version. Then the caller passes the version in
with the write operation and the operation will fail if another caller has
updated the resource's version in the meantime.

```go
err := rw.LookupId(ctx, &user)

user.Name = "Alice"
rowsAffected, err = rw.Update(ctx, 
    &user, 
    []string{"Name"}, 
    nil, 
    dbw.WithVersion(&user.Version))

if err != nil && error.Is(err, dbw.ErrRecordNotFound) {
    // update failed because the row wasn't found because  
    // either it was deleted, or updated by another caller 
    // after it was read earlier in this example
}
```