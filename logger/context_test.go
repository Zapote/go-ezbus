package logger

import (
	"context"
	"log/slog"
	"testing"
)

type ctxKey struct{}

// recordingHandler keeps the context and the record of the last Handle call.
type recordingHandler struct {
	ctx    context.Context
	record slog.Record
	calls  int
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *recordingHandler) Handle(ctx context.Context, r slog.Record) error {
	h.ctx, h.record, h.calls = ctx, r, h.calls+1
	return nil
}
func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }

func TestContextAndAttrsReachTheHandler(t *testing.T) {
	h := &recordingHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	defer slog.SetDefault(prev)
	SetLevel(InfoLevel)

	ctx := context.WithValue(context.Background(), ctxKey{}, "value")
	ErrorContext(ctx, "attempt failed", "attempt", 3)

	if h.calls != 1 {
		t.Fatalf("handler called %d times", h.calls)
	}
	if h.ctx.Value(ctxKey{}) != "value" {
		t.Error("the context did not reach the handler")
	}
	if h.record.Message != "attempt failed" || h.record.Level != slog.LevelError {
		t.Errorf("record = %q %v", h.record.Message, h.record.Level)
	}
	var attempt int64
	h.record.Attrs(func(a slog.Attr) bool {
		if a.Key == "attempt" {
			attempt = a.Value.Int64()
		}
		return true
	})
	if attempt != 3 {
		t.Errorf("attempt = %d", attempt)
	}
}

func TestContextVariantsRespectTheLevel(t *testing.T) {
	h := &recordingHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	defer slog.SetDefault(prev)
	SetLevel(WarnLevel)
	defer SetLevel(InfoLevel)

	DebugContext(context.Background(), "quiet")
	InfoContext(context.Background(), "quiet")
	WarnContext(context.Background(), "loud")

	if h.calls != 1 || h.record.Message != "loud" {
		t.Errorf("calls = %d, last = %q", h.calls, h.record.Message)
	}
}
