# Connecting
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

dbw has tested official support for SQLite and Postgres.  You can also use it to connect
to any database that has a Gorm V2 driver which has official support for: SQLite, Postgres,
MySQL and SQL Server. 

## SQLite
```go
import(
    "github.com/hashicorp/go-dbw"
)

func main() {
    db, err := dbw.Open(dbw.Sqlite, "dbw.db")    
}
```

## Postgres
```go
import(
    "github.com/hashicorp/go-dbw"
)

func main() {
    dsn := "postgresql://go_db:go_db@localhost:9920/go_db?sslmode=disable"
    db, err := dbw.Open(dbw.Postgres, dsn)    
}
```

## Any gorm v2 driver or an existing connection
```go
import(
    "database/sql"
    "github.com/hashicorp/go-dbw"
    "gorm.io/gorm"
)

func main() {
    dsn := "postgresql://go_db:go_db@localhost:9920/go_db?sslmode=disable"
    sqlDB, err := sql.Open("mysql", dsn)
    db, err := dbw.OpenWith(mysql.New(mysql.Config{
        Conn: sqlDB,
    }))
}
```

## Connection Pooling

```go
import(
    "github.com/hashicorp/go-dbw"
)

func main() {
    dsn := "postgresql://go_db:go_db@localhost:9920/go_db?sslmode=disable"
    db, err := dbw.Open(dbw.Postgres, dsn, 
        dbw.WithMaxConnections(20),
        dbw.WithMinConnections(2),
    )    
    sqlDB, err = db.SqlDB()
    sqlDB.SetConnMaxLifetime(time.Hour)
    sqlDB.SetMaxIdleConns(10)
}
```