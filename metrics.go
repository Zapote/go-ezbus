package ezbus

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "github.com/zapote/go-ezbus"

// Operation and outcome values used as metric attributes.
const (
	operationSend    = "send"
	operationPublish = "publish"
	operationProcess = "process"

	outcomeOK         = "ok"
	outcomeError      = "error"
	outcomeErrorQueue = "error_queue"
	outcomeDiscarded  = "discarded"
)

// The instruments follow the OpenTelemetry semantic conventions for
// messaging where one exists, so a dashboard made for another client reads
// the same. They are created once, against the global meter provider, and
// stay no-ops until the service installs one.
var (
	sentMessages      metric.Int64Counter
	consumedMessages  metric.Int64Counter
	operationDuration metric.Float64Histogram
	processDuration   metric.Float64Histogram
	processAttempts   metric.Int64Histogram
)

func init() {
	m := otel.Meter(meterName)
	var err error

	sentMessages, err = m.Int64Counter("messaging.client.sent.messages",
		metric.WithUnit("{message}"),
		metric.WithDescription("Messages sent or published by the bus, by destination and outcome."))
	reportErr(err)

	consumedMessages, err = m.Int64Counter("messaging.client.consumed.messages",
		metric.WithUnit("{message}"),
		metric.WithDescription("Messages taken off the queue and handled, by message name and outcome."))
	reportErr(err)

	operationDuration, err = m.Float64Histogram("messaging.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Time to send or publish a message."))
	reportErr(err)

	processDuration, err = m.Float64Histogram("messaging.process.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Time to handle a message, all attempts included."))
	reportErr(err)

	processAttempts, err = m.Int64Histogram("ezbus.process.attempts",
		metric.WithUnit("{attempt}"),
		metric.WithDescription("Attempts it took to handle a message. Five means it ended on the error queue."))
	reportErr(err)
}

func reportErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}

// recordSend counts a send or publish and how long it took.
func recordSend(ctx context.Context, operation string, destination string, start time.Time, err error) {
	outcome := outcomeOK
	if err != nil {
		outcome = outcomeError
	}

	attrs := metric.WithAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.operation.name", operation),
		attribute.String("messaging.destination.name", destination),
		attribute.String("ezbus.outcome", outcome),
	)

	sentMessages.Add(ctx, 1, attrs)
	operationDuration.Record(ctx, time.Since(start).Seconds(), attrs)
}

// recordProcess counts a handled message, how long it took over all
// attempts and how many attempts it took.
func recordProcess(ctx context.Context, queue string, messageName string, start time.Time, attempts int, outcome string) {
	attrs := metric.WithAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.operation.name", operationProcess),
		attribute.String("messaging.destination.name", queue),
		attribute.String("ezbus.message_name", messageName),
		attribute.String("ezbus.outcome", outcome),
	)

	consumedMessages.Add(ctx, 1, attrs)
	processDuration.Record(ctx, time.Since(start).Seconds(), attrs)
	processAttempts.Record(ctx, int64(attempts), attrs)
}
