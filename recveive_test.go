package ezbus

import (
	"errors"
	"testing"

	"gotest.tools/assert"
)

func TestRetryRunsThreeAttempsOnError(t *testing.T) {
	n := 0

	err := receive("handle", func() error {
		n++
		return errors.New("this wont work")
	}, 3)

	assert.Equal(t, n, 3)
	assert.Check(t, err != nil)
}

func TestRetryRunsOnlyOneAttempOnPanic(t *testing.T) {
	n := 0

	err := receive("handle", func() error {
		n++
		panic("this wont work")
	}, 3)

	assert.Equal(t, n, 1)
	assert.Check(t, err != nil)
}

func TestRetryRunsOnlyOneAttempOnHandlerNotFound(t *testing.T) {
	n := 0

	err := receive("handle", func() error {
		n++
		return HandlerNotFoundErr{}
	}, 3)

	assert.Equal(t, n, 1)
	assert.Check(t, err != nil)
}

func TestRetryOnlyRunsOnceWhenSuccess(t *testing.T) {
	n := 0

	err := receive("handle", func() error {
		n++
		return nil
	}, 3)

	assert.Equal(t, n, 1)
	assert.Check(t, err == nil)
}
