package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"gotest.tools/assert"
)

func capture(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(previous)
		SetLevel(InfoLevel)
	})

	return buf
}

func line(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	l := map[string]any{}
	if err := json.Unmarshal(buf.Bytes(), &l); err != nil {
		t.Fatalf("not a json log line: %s (%s)", err, buf.String())
	}

	return l
}

func TestWritesThroughSlog(t *testing.T) {
	buf := capture(t)

	Info("bus started")

	l := line(t, buf)
	assert.Equal(t, "INFO", l["level"])
	assert.Equal(t, "bus started", l["msg"])
}

func TestFormatsArguments(t *testing.T) {
	buf := capture(t)

	Errorf("failed to handle %s after %d attempts", "OrderPlaced", 5)

	l := line(t, buf)
	assert.Equal(t, "ERROR", l["level"])
	assert.Equal(t, "failed to handle OrderPlaced after 5 attempts", l["msg"])
}

func TestLeavesPercentAloneWithoutArguments(t *testing.T) {
	buf := capture(t)

	Warn("queue is 90% full")

	l := line(t, buf)
	assert.Equal(t, "queue is 90% full", l["msg"])
}

func TestRespectsLevel(t *testing.T) {
	buf := capture(t)
	SetLevel(WarnLevel)

	Info("this one is filtered")

	assert.Equal(t, 0, buf.Len())
}
