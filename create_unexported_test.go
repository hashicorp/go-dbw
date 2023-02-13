// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dbw

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NonCreatableFields(t *testing.T) {
	// do not run with t.Parallel()
	assert := assert.New(t)
	nonUpdateFields = atomic.Value{}
	got := NonCreatableFields()
	assert.Equal(got, []string{})

	InitNonCreatableFields([]string{"Foo"})
	got = NonCreatableFields()
	assert.Equal(got, []string{"Foo"})
}
