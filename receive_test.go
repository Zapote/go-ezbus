package ezbus

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"gotest.tools/v3/assert"
)

func TestRetryRunsThreeAttempsOnError(t *testing.T) {
	n := 0

	_, err := receive(context.Background(), "handle", func() error {
		n++
		return errors.New("this wont work")
	}, 3)

	assert.Equal(t, n, 3)
	assert.Check(t, err != nil)
}

func TestRetryRunsOnlyOneAttempOnPanic(t *testing.T) {
	n := 0

	_, err := receive(context.Background(), "handle", func() error {
		n++
		panic("this wont work")
	}, 3)

	assert.Equal(t, n, 1)
	assert.Check(t, err != nil)
}

func TestRetryRunsOnlyOneAttempOnHandlerNotFound(t *testing.T) {
	n := 0

	_, err := receive(context.Background(), "handle", func() error {
		n++
		return HandlerNotFoundErr{}
	}, 3)

	assert.Equal(t, n, 1)
	assert.Check(t, err != nil)
}

func TestRetryOnlyRunsOnceWhenSuccess(t *testing.T) {
	n := 0

	_, err := receive(context.Background(), "handle", func() error {
		n++
		return nil
	}, 3)

	assert.Equal(t, n, 1)
	assert.Check(t, err == nil)
}

func TestRetryReturnsTheAttemptCount(t *testing.T) {
	n := 0

	attempts, err := receive(context.Background(), "handle", func() error {
		n++
		if n < 3 {
			return errors.New("not yet")
		}
		return nil
	}, 5)

	assert.Equal(t, 3, attempts)
	assert.Check(t, err == nil)
}

type contextRecorder struct {
	slog.Handler
	ctx context.Context
}

func (h *contextRecorder) Handle(ctx context.Context, r slog.Record) error {
	h.ctx = ctx
	return h.Handler.Handle(ctx, r)
}

func TestRetryLogsWithTheContext(t *testing.T) {
	rec := &contextRecorder{Handler: slog.NewTextHandler(io.Discard, nil)}
	prev := slog.Default()
	slog.SetDefault(slog.New(rec))
	defer slog.SetDefault(prev)

	ctx := context.WithValue(context.Background(), ctxKey{}, "value")
	receive(ctx, "handle", func() error { return errors.New("no luck") }, 2)

	if rec.ctx == nil || rec.ctx.Value(ctxKey{}) != "value" {
		t.Error("the retry line was not logged with the message context")
	}
}
