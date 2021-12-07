# Transactions
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

`dbw` supports transactions via `dbw.DoTx(...)`  `DoTx` supports any backoff
strategy that implements the `Backoff` interface.  There are two backoffs
provided by the package: `ConstBackOff` and `ExpBackoff`.

```go
// Example with ExpBackoff
retryErrFn := func(_ error) bool { return true }
_, err = rw.DoTx(
    context.Background(),
    func(_ error) bool { return true }, // retry all errors
    3,                                  // three retries
    ExpBackoff{},                       // exponential backoff
    func(w Writer) error {
        // the TxHandler updates the user's name
        _, err := w.Update(context.Background(), 
            user, 
            []string{"Name"}, 
            nil,
            dbw.WithVersion(&user.Version),
        )
        if err != nil {
            return err
        }
    },
)
if err != nil {
    // handle errors from the transaction...
}
```

You can also control the transaction yourself using `Begin`, `Rollback`, and
`Commit`.

```go
// begin a transaction
tx, err := rw.Begin(ctx)

// do some database operations like creating a resource
if err := tx.Create(...); err != nil {

    // rollback the transaction if you 
    if err := tx.Rollback(ctx); err != nil {
        // you'll need to handle rollback errors... perhaps via retry.
    }
}

// commit the transaction if there are not errors
if err := tx.Commit(ctx); err != nil {
    // handle commit errors
}
```