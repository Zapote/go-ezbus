package ezbus

import (
	"testing"

	"github.com/zapote/go-ezbus/assert"
)

func TestRetryRunsThreeAttempsWhenPanic(t *testing.T) {
	n := 0

	err := retry(func() {
		n++
		panic("this wont work")
	}, 3)

	assert.IsEqual(t, n, 3)
	assert.IsNotNil(t, err)
}

func TestRetryOnlyRunsOnceWhenSuccess(t *testing.T) {
	n := 0

	err := retry(func() {
		n++
	}, 3)

	assert.IsEqual(t, n, 1)
	assert.IsNil(t, err)
}
