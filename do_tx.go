package dbw

import (
	"context"
	"fmt"
	"time"
)

// DoTx will wrap the Handler func passed within a transaction with retries
// you should ensure that any objects written to the db in your TxHandler are retryable, which
// means that the object may be sent to the db several times (retried), so
// things like the primary key may need to be reset before retry.
func (w *RW) DoTx(ctx context.Context, retryErrorsMatchingFn func(error) bool, retries uint, backOff Backoff, Handler TxHandler) (RetryInfo, error) {
	const op = "dbw.DoTx"
	if w.underlying == nil {
		return RetryInfo{}, fmt.Errorf("%s: missing underlying db: %w", op, ErrInvalidParameter)
	}
	if backOff == nil {
		return RetryInfo{}, fmt.Errorf("%s: missing backoff: %w", op, ErrInvalidParameter)
	}
	if Handler == nil {
		return RetryInfo{}, fmt.Errorf("%s: missing handler: %w", op, ErrInvalidParameter)
	}
	if retryErrorsMatchingFn == nil {
		return RetryInfo{}, fmt.Errorf("%s: missing retry errors matching function: %w", op, ErrInvalidParameter)
	}
	info := RetryInfo{}
	for attempts := uint(1); ; attempts++ {
		if attempts > retries+1 {
			return info, fmt.Errorf("%s: too many retries: %d of %d: %w", op, attempts-1, retries+1, ErrMaxRetries)
		}

		// step one of this, start a transaction...
		newTx := w.underlying.wrapped.WithContext(ctx)
		newTx = newTx.Begin()

		rw := &RW{underlying: &DB{newTx}}
		if err := Handler(rw, rw); err != nil {
			if err := newTx.Rollback().Error; err != nil {
				return info, fmt.Errorf("%s: %w", op, err)
			}
			if retry := retryErrorsMatchingFn(err); retry {
				d := backOff.Duration(attempts)
				info.Retries++
				info.Backoff = info.Backoff + d
				time.Sleep(d)
				continue
			}
			return info, fmt.Errorf("%s: %w", op, err)
		}

		if err := newTx.Commit().Error; err != nil {
			if err := newTx.Rollback().Error; err != nil {
				return info, fmt.Errorf("%s: %w", op, err)
			}
			return info, fmt.Errorf("%s: %w", op, err)
		}
		return info, nil // it all worked!!!
	}
}
