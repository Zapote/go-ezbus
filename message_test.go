package ezbus

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

type ctxKey struct{}

func TestMessageWithoutContext(t *testing.T) {
	m := NewMessage(map[string]string{}, []byte("body"))

	assert.Equal(t, context.Background(), m.Context())
}

func TestMessageWithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey{}, "value")

	m := NewMessage(map[string]string{}, []byte("body")).WithContext(ctx)

	assert.Equal(t, "value", m.Context().Value(ctxKey{}))
}

func TestWithContextLeavesOriginalAlone(t *testing.T) {
	m := NewMessage(map[string]string{}, []byte("body"))

	m.WithContext(context.WithValue(context.Background(), ctxKey{}, "value"))

	assert.Equal(t, nil, m.Context().Value(ctxKey{}))
}
