package ezbus

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/zapote/go-ezbus/headers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"gotest.tools/v3/assert"
)

var (
	traceID = trace.TraceID{0x4b, 0xf9, 0x2f, 0x35, 0x77, 0xb3, 0x4d, 0xa6, 0xa3, 0xce, 0x92, 0x9d, 0x0e, 0x0e, 0x47, 0x36}
	spanID  = trace.SpanID{0x00, 0xf0, 0x67, 0xaa, 0x0b, 0xa9, 0x02, 0xb7}
)

func tracingContext() context.Context {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	return trace.ContextWithSpanContext(context.Background(), sc)
}

func TestSendContextWritesTraceparent(t *testing.T) {
	b.SendContext(tracingContext(), "queue-name", msg)

	m := broker.sentMessage.(Message)
	assert.Check(t, strings.HasPrefix(m.Headers["traceparent"], "00-"+traceID.String()+"-"),
		"traceparent was %q", m.Headers["traceparent"])
}

func TestSendWithoutTraceWritesNoTraceparent(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	b.Send("queue-name", msg)

	m := broker.sentMessage.(Message)
	_, ok := m.Headers["traceparent"]
	assert.Check(t, !ok, "traceparent should not be set without a trace")
}

func TestHandleContinuesTheTraceFromTheHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	var wg sync.WaitGroup
	wg.Add(1)
	var handled trace.SpanContext

	rtr.Handle("FakeMessage", func(m Message) error {
		handled = trace.SpanContextFromContext(m.Context())
		defer wg.Done()
		return nil
	})

	go b.Go()
	defer b.Stop()
	broker.invokeWith(map[string]string{
		"traceparent": "00-" + traceID.String() + "-" + spanID.String() + "-01",
	})
	wg.Wait()

	assert.Equal(t, traceID.String(), handled.TraceID().String())
}

// recordSpans installs a tracer provider that keeps every ended span, for
// the rest of the test.
func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	sr := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)))
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTracerProvider(noop.NewTracerProvider()) })

	return sr
}

func processSpan(t *testing.T, sr *tracetest.SpanRecorder) sdktrace.ReadOnlySpan {
	t.Helper()

	for _, s := range sr.Ended() {
		if s.Name() == "process fake-broker" {
			return s
		}
	}

	t.Fatal("no process span ended")
	return nil
}

func TestHandleRecordsEachFailedAttempt(t *testing.T) {
	sr := recordSpans(t)
	rtr.Handle("FakeMessage", func(m Message) error {
		return errors.New("no luck")
	})

	go b.Go()
	defer b.Stop()
	broker.invokeWith(map[string]string{})

	span := processSpan(t, sr)
	assert.Equal(t, 5, len(span.Events()))
	assert.Equal(t, "exception", span.Events()[0].Name)
	assert.Equal(t, codes.Error, span.Status().Code)
}

func TestMessageOnErrorQueueCarriesTheFailedAttempt(t *testing.T) {
	sr := recordSpans(t)
	rtr.Handle("FakeMessage", func(m Message) error {
		return errors.New("no luck")
	})

	go b.Go()
	defer b.Stop()
	broker.invokeWith(map[string]string{})

	sc := processSpan(t, sr).SpanContext()
	m := broker.sentMessage.(Message)
	assert.Equal(t, "fake-broker-error", broker.sentDst)
	assert.Equal(t, "00-"+sc.TraceID().String()+"-"+sc.SpanID().String()+"-01", m.Headers["traceparent"])
}

func TestRerunFromErrorQueueStartsANewTraceLinkedToTheOldOne(t *testing.T) {
	sr := recordSpans(t)
	rtr.Handle("FakeMessage", func(m Message) error {
		return nil
	})

	go b.Go()
	defer b.Stop()
	broker.invokeWith(map[string]string{
		"traceparent": "00-" + traceID.String() + "-" + spanID.String() + "-01",
		headers.Error: "no luck",
	})

	span := processSpan(t, sr)
	assert.Check(t, span.SpanContext().TraceID() != traceID, "a rerun should start a new trace")
	assert.Equal(t, 1, len(span.Links()))
	assert.Equal(t, traceID, span.Links()[0].SpanContext.TraceID())
	assert.Equal(t, codes.Unset, span.Status().Code)
}
