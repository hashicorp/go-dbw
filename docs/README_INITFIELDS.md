# NonCreatable and NonUpdatable fields
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` provides a set of functions which allows you to define sets of fields
which cannot be set using
[RW.Create(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW.Create) or
updated via
[RW.Update(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW.Update). To 
be clear, errors are not raised if you mistakenly try to set/update these
fields, but rather `dbw` quietly removes the set/update of these fields before
generating the sql to send along to the database.

For more details see:
* [InitNonCreatableFields](https://pkg.go.dev/github.com/hashicorp/go-dbw#InitNonCreatableFields)
* [InitNonUpdatableFields](https://pkg.go.dev/github.com/hashicorp/go-dbw#InitNonUpdatableFields)
* [NonCreatableFields](https://pkg.go.dev/github.com/hashicorp/go-dbw#NonCreatableFields)
* [NonUpdatableFields](https://pkg.go.dev/github.com/hashicorp/go-dbw#NonUpdatableFields)

```go
// initialize fields which cannot be set during creation
dbw.InitNonCreatableFields([]string{"CreateTime", "UpdateTime"})
// read the current set of non-creatable fields
fields := dbw.NonCreatableFields() 

// initialize fields which cannot be updated
dbw.InitNonUpdatableFields([]string{"PublicId", "CreateTime", "UpdateTime"})
// read the current set of non-updatable fields
fields = dbw.NonUpdatableFields() 
```