package ezbus

import (
	"errors"
	"testing"

	"github.com/zapote/go-ezbus/assert"
)

func TestRetryRunsThreeAttempsWhenPanic(t *testing.T) {
	n := 0

	err := retry(func() error {
		n++
		return errors.New("this wont work")
	}, 3)

	assert.IsEqual(t, n, 3)
	assert.IsNotNil(t, err)
}

func TestRetryOnlyRunsOnceWhenSuccess(t *testing.T) {
	n := 0

	err := retry(func() error {
		n++
		return nil
	}, 3)

	assert.IsEqual(t, n, 1)
	assert.IsNil(t, err)
}
