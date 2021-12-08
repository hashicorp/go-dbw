# Hooks
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

dbw provides two options for write operations which give callers hooks before
and after the write operations:
* [WithBeforeWrite(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithBeforeWrite) 
* [WithAfterWrite(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithAfterWrite)

```go 
beforeFn := func(_ interface{}) error {	
    return nil // always succeed for this example
}
afterFn := func(_ interface{}, _ int) error { 
    return nil // always succeed for this example
}

rw.Create(ctx, 
    &user, 
    dbw.WithBeforeWrite(beforeFn), 
    dbw.WithAfterWrite(afterFn),
)

```

