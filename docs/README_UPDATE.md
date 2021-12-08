# Update
[![Go Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

[Update(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW.Update) requires
the resource to be updated with its fields set that the caller wants updated. 

A `fieldMask` is optional and provides paths for fields that
should be updated. 

A `setToNullPaths` is optional and provides paths for the fields that should be
set to null.

Either a `fieldMaskPaths` or `setToNullPaths` must be provided and they must not intersect.

The caller is responsible for the transaction life cycle of the writer and if an
error is returned the caller must decide what to do with the transaction, which
almost always should be to rollback.  Update returns the number of rows updated.

There a lots of supported options:
[WithBeforeWrite](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithBeforeWrite),
[WithAfterWrite](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithAfterWrite),
[WithWhere](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithWhere),
[WithDebug](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithDebug), and 
[WithVersion](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithVersion). 

If [WithVersion](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithVersion) is
used, then the update will include the version number in the
update where clause, which basically makes the update use optimistic locking and
the update will only succeed if the existing rows version matches the
[WithVersion](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithVersion)
option. Zero is not a valid value for the
[WithVersion](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithVersion) option
and will return an error. 

[WithWhere](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithWhere) allows
specifying an additional constraint on the operation in
addition to the PKs. 

[WithDebug](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithDebug) will turn
on debugging for the update call. 

### Simple update [WithVersion](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithVersion) example
```go
user.Name = "Alice"
rowsAffected, err = rw.Update(ctx, 
    &user, 
    []string{"Name"}, 
    nil, 
    dbw.WithVersion(&user.Version))
```

### Update with setToNullPaths and `WithVersion` example
```go
user.Name = "Alice"
rowsAffected, err = rw.Update(ctx, 
    &user, 
    nil, 
    []string{"Name"}, 
    dbw.WithVersion(&user.Version))
```