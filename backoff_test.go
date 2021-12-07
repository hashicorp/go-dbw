package dbw

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstBackoff_Duration(t *testing.T) {
	tests := []struct {
		name    string
		b       ConstBackoff
		attempt int
		want    time.Duration
	}{
		{
			name:    "one",
			b:       ConstBackoff{DurationMs: 2},
			attempt: 1,
			want:    time.Millisecond * 2,
		},
		{
			name:    "two",
			b:       ConstBackoff{DurationMs: 2},
			attempt: 2,
			want:    time.Millisecond * 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got := tt.b.Duration(uint(tt.attempt))
			assert.Equal(tt.want, got)
		})
	}
}

func TestExpBackoff_Duration(t *testing.T) {
	tests := []struct {
		name     string
		b        ExpBackoff
		attempt  int
		want     time.Duration
		wantRand bool
	}{
		{
			name:    "one",
			b:       ExpBackoff{testRand: 1},
			attempt: 1,
			want:    time.Millisecond * 15,
		},
		{
			name:    "two",
			b:       ExpBackoff{testRand: 2},
			attempt: 1,
			want:    time.Millisecond * 25,
		},
		{
			name:     "rand",
			b:        ExpBackoff{},
			attempt:  1,
			wantRand: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got := tt.b.Duration(uint(tt.attempt))
			if tt.wantRand {
				assert.NotZero(got)
				return
			}
			assert.Equal(tt.want, got)
		})
	}
}
