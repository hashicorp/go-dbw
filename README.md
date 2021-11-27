# dbw package
[![Go Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` is a database wrapper that supports connecting and using any database with a
[GORM](https://github.com/go-gorm/gorm) driver.  It's intent is to completely
encapsulate an application's access to it's database with the exception of
migrations.    

`dbw` is intentionally not an ORM and it removes typical ORM abstractions like
"advanced query building", associations and migrations.  

This is not to say you can't easily use dbw for complicated queries, it's just
that dbw doesn't try to reinvent sql by providing some sort of pattern for
building them with functions. Of course, dbw also provides lookup/search
functions when you simply need to read resources from the database.

`dbw` strives to make CRUD for database resources fairly trivial.  Even supporting
"on conflict" for its create function.  dbw also allows you to opt out of its
CRUD functions and use exec, query and scan rows directly.  You may want to
carefully weigh when it's appropriate to use exec and query directly, since
it's likely that each time you use them you're leaking a bit of your
database schema into your application's domain. 

* [Usage highlights](./README_USAGE.md)
* [Declaring Models](./README_MODELS.md)
* [Connecting to a Database](./README_OPEN.md)
* [NonCreatable and NonUpdatable](./README_INITFIELDS.md)
* [Readers and Writers](./README_RW.md)
* [Create](./README_CREATE.md)
* [Read](./README_READ.md)
* [Update](./README_UPDATE.md)
* [Delete](./README_DELETE.md)
* [Queries](./README_QUERY.md)
* [Transactions](./README_TX.md)
* [Hooks](./README_HOOKS.md)
* [Optimistic locking for write operations](./README_LOCKs.md)
* [Debug output](./README_DEBUG.md)