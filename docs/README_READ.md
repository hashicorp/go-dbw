# Reading
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

dbw provides a few ways to read data from the database.

```go
// Get the user with either a PublicId, PrivateId or 
// primary keys matching the given users
rw.LookupId(ctx, &user)

// Get the user with a public_id matching the given user.
rw.LookupByPublicId(ctx, &user)

// Get the first user matching the where clause
rw.LookupWhere(ctx, 
    &user, 
    "public_id = @id", 
    sql.Named("id", "1"),
)

// Get all the users matching the where clause
rw.SearchWhere(ctx, 
    &users, 
    "public_id in(@ids)", 
    sql.Named("ids", []string{"1", "2"}),
)
```