package dbw

import (
	"math"
	"math/rand"
	"time"
)

type Backoff interface {
	Duration(attemptNumber uint) time.Duration
}

type ConstBackoff struct {
	DurationMs time.Duration
}

func (b ConstBackoff) Duration(attempt uint) time.Duration {
	return time.Millisecond * time.Duration(b.DurationMs)
}

type ExpBackoff struct {
	testRand float64
}

func (b ExpBackoff) Duration(attempt uint) time.Duration {
	var r float64
	switch {
	case b.testRand > 0:
		r = b.testRand
	default:
		r = rand.Float64()
	}
	return time.Millisecond * time.Duration(math.Exp2(float64(attempt))*5*(r+0.5))
}
