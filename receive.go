package ezbus

import (
	"context"
	"fmt"

	"github.com/zapote/go-ezbus/logger"
)

// receive runs fn up to attempts times, until it succeeds. Every failed
// attempt is recorded on the span in ctx, so the retries show in the trace.
func receive(ctx context.Context, messageName string, fn func() error, attempts int) (err error) {
	attempt := 0

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered from panic: %v", r)
			recordAttempt(ctx, attempt, err)
		}
	}()

	for i := 0; i < attempts; i++ {
		attempt = i + 1
		if err = fn(); err != nil {
			switch v := err.(type) {
			case HandlerNotFoundErr:
				return v
			default:
				logger.Errorf("Attempt #%d, message '%s' failed: %s", attempt, messageName, err.Error())
				recordAttempt(ctx, attempt, err)
				continue
			}
		}
		return nil
	}
	return err
}
