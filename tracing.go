package ezbus

import (
	"context"

	"github.com/zapote/go-ezbus/headers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/zapote/go-ezbus"

// startPublish starts a producer span and writes the trace context into h.
// Both are no-ops until the service sets up a tracer provider and a propagator.
func startPublish(ctx context.Context, destination string, h map[string]string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "publish "+destination,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation.name", "publish"),
			attribute.String("messaging.destination.name", destination),
		))

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(h))

	return ctx, span
}

// startProcess starts a consumer span, with the span that published the
// message as its parent when h carries a trace context.
//
// A message that comes back from the error queue still carries the trace
// context of the attempt that failed, possibly days ago. The error header
// tells it apart, and it gets a new trace with a link to the old one
// instead of continuing it.
func startProcess(queue string, messageName string, h map[string]string) (context.Context, trace.Span) {
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(h))

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation.name", "process"),
			attribute.String("messaging.destination.name", queue),
			attribute.String("ezbus.message_name", messageName),
		),
	}

	if _, rerun := h[headers.Error]; rerun {
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			opts = append(opts, trace.WithLinks(trace.Link{SpanContext: sc}))
		}
		opts = append(opts, trace.WithAttributes(attribute.Bool("ezbus.rerun", true)))
		ctx = context.Background()
	}

	return otel.Tracer(tracerName).Start(ctx, "process "+queue, opts...)
}

// recordAttempt records a failed attempt at handling a message as an
// exception event on the consumer span, numbered so the retries can be
// told apart.
func recordAttempt(ctx context.Context, attempt int, err error) {
	trace.SpanFromContext(ctx).RecordError(err, trace.WithAttributes(attribute.Int("ezbus.attempt", attempt)))
}

// markErrorQueued writes the consumer span into h, so that a rerun from the
// error queue links to the attempt that failed rather than to the publish
// that started it.
func markErrorQueued(ctx context.Context, h map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(h))
}

// failSpan marks the span as failed and records the error on it.
func failSpan(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// failAfterAttempts marks the consumer span as failed. The errors are
// already on it, one event per attempt.
func failAfterAttempts(span trace.Span, err error) {
	span.SetStatus(codes.Error, err.Error())
}
