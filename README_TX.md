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

```