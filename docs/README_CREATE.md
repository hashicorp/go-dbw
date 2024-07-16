# Create
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

## [RW.Create(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW.Create) example with one item
```go
id, err := dbw.NewId("u")

user := TestUser{PublicId: id, Name: "Alice"}

var rowsAffected int64
err = rw.Create(ctx, &user, dbw.WithRowsAffected(&rowsAffected))  
```
## [RW.CreateItems(...)](https://pkg.go.dev/github.com/hashicorp/go-dbw#RW.CreateItems) example with multiple items
```go
var rowsAffected int64
err = rw.CreateItems(ctx, []*dbtest.TestUser{&user1, &user2}, dbw.WithRowsAffected(&rowsAffected))  
```


## [OnConflict](https://pkg.go.dev/github.com/hashicorp/go-dbw#WithOnConflict) upsert example

Upserts via a variety of conflict targets and actions are supported.

```go
// set columns
onConflict := dbw.OnConflict{
    Target: dbw.Columns{"public_id"},
    Action: dbw.SetColumns([]string{"name"}),
}
rw.Create(ctx, &user, dbw.WithConflict(&onConflict))
```

```go
// set columns and column values
onConflict := dbw.OnConflict{
	Target: dbw.Columns{"public_id"},
}
cv := dbw.SetColumns([]string{"name"})
cv = append(
    cv,
	dbw.SetColumnValues(map[string]interface{}{
	"email":        "alice@gmail.com",
	"phone_number": dbw.Expr("NULL"),
})...)
onConflict.Action = cv
rw.Create(ctx, &user, dbw.WithConflict(&onConflict))
```

```go
// do nothing
onConflict := dbw.OnConflict{
    Target: dbw.Columns{"public_id"},
    Action: dbw.DoNothing(true),
}
rw.Create(ctx, &user, dbw.WithConflict(&onConflict))
```

```go
// on constraint
onConflict := dbw.OnConflict{
    Target: dbw.Constraint("db_test_user_pkey"),
    Action: dbw.SetColumns([]string{"name"}),
}
rw.Create(ctx, &user, dbw.WithConflict(&onConflict))
```

```go
// set columns combined with WithVersion
onConflict := dbw.OnConflict{
    Target: dbw.Columns{"public_id"},
	Action: dbw.SetColumns([]string{"name"}),
}
version := uint32(1)
rw.Create(ctx, &user, dbw.WithConflict(&onConflict), dbw.WithVersion(&version))
```

