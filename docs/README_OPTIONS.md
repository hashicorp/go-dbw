# Options
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` supports variadic 
[Option](https://pkg.go.dev/github.com/hashicorp/go-dbw#Option) function
parameters for the vast majority of its operations.  See the [dbw package
docs](https://pkg.go.dev/github.com/hashicorp/go-dbw) for more
information about which
[Option](https://pkg.go.dev/github.com/hashicorp/go-dbw#Option) functions are
supported for each operation.


```go
// just one example of variadic options: an update 
// using WithVersion and WithDebug options
rw.Update(ctx, &user, dbw.WithVersion(10), dbw.WithDebug(true))
```