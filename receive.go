package ezbus

import (
	"context"
	"fmt"

	"github.com/zapote/go-ezbus/logger"
)

// receive runs fn up to limit times, until it succeeds, and returns how
// many attempts it took. Every failed attempt is recorded on the span in
// ctx, so the retries show in the trace.
func receive(ctx context.Context, messageName string, fn func() error, limit int) (attempt int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered from panic: %v", r)
			recordAttempt(ctx, attempt, err)
		}
	}()

	for i := 0; i < limit; i++ {
		attempt = i + 1
		if err = fn(); err != nil {
			switch v := err.(type) {
			case HandlerNotFoundErr:
				return attempt, v
			default:
				logger.Errorf("Attempt #%d, message '%s' failed: %s", attempt, messageName, err.Error())
				recordAttempt(ctx, attempt, err)
				continue
			}
		}
		return attempt, nil
	}
	return attempt, err
}
