// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dbw

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NonUpdatableFields(t *testing.T) {
	// do not run with t.Parallel()
	assert := assert.New(t)
	nonUpdateFields = atomic.Value{}
	got := NonUpdatableFields()
	assert.Equal(got, []string{})

	InitNonUpdatableFields([]string{"Foo"})
	got = NonUpdatableFields()
	assert.Equal(got, []string{"Foo"})
}
