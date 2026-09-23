package ezbus

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/zapote/go-ezbus/headers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"gotest.tools/v3/assert"
)

var (
	metricsOnce   sync.Once
	metricsReader *sdkmetric.ManualReader
)

// collectMetrics installs a meter provider once, with delta temporality so
// that a collection holds only what happened since the previous one, and
// discards whatever earlier tests recorded. The returned function collects.
func collectMetrics(t *testing.T) func() metricdata.ResourceMetrics {
	t.Helper()

	metricsOnce.Do(func() {
		metricsReader = sdkmetric.NewManualReader(sdkmetric.WithTemporalitySelector(
			func(sdkmetric.InstrumentKind) metricdata.Temporality { return metricdata.DeltaTemporality }))
		otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricsReader)))
	})

	collect := func() metricdata.ResourceMetrics {
		var rm metricdata.ResourceMetrics
		assert.NilError(t, metricsReader.Collect(context.Background(), &rm))
		return rm
	}
	collect()

	return collect
}

func sumOf(rm metricdata.ResourceMetrics, name string, want ...attribute.KeyValue) int64 {
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				continue
			}
			for _, dp := range sum.DataPoints {
				if hasAll(dp.Attributes, want) {
					total += dp.Value
				}
			}
		}
	}
	return total
}

func histogramOf[N int64 | float64](rm metricdata.ResourceMetrics, name string, want ...attribute.KeyValue) (count uint64, sum N) {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			h, ok := m.Data.(metricdata.Histogram[N])
			if !ok {
				continue
			}
			for _, dp := range h.DataPoints {
				if hasAll(dp.Attributes, want) {
					count += dp.Count
					sum += dp.Sum
				}
			}
		}
	}
	return count, sum
}

func hasAll(set attribute.Set, want []attribute.KeyValue) bool {
	for _, kv := range want {
		v, ok := set.Value(kv.Key)
		if !ok || v.AsString() != kv.Value.AsString() {
			return false
		}
	}
	return true
}

func TestSendIsCountedAndTimed(t *testing.T) {
	collect := collectMetrics(t)

	b.Send("queue-name", msg)

	rm := collect()
	want := []attribute.KeyValue{
		attribute.String("messaging.operation.name", "send"),
		attribute.String("messaging.destination.name", "queue-name"),
		attribute.String("ezbus.outcome", "ok"),
	}
	assert.Equal(t, int64(1), sumOf(rm, "messaging.client.sent.messages", want...))
	count, _ := histogramOf[float64](rm, "messaging.client.operation.duration", want...)
	assert.Equal(t, uint64(1), count)
}

func TestPublishIsCountedByMessageType(t *testing.T) {
	collect := collectMetrics(t)

	b.Publish(msg)

	rm := collect()
	assert.Equal(t, int64(1), sumOf(rm, "messaging.client.sent.messages",
		attribute.String("messaging.operation.name", "publish"),
		attribute.String("messaging.destination.name", "FakeMessage"),
		attribute.String("ezbus.outcome", "ok")))
}

func TestHandledMessageIsCountedWithOneAttempt(t *testing.T) {
	collect := collectMetrics(t)
	rtr.Handle("FakeMessage", func(m Message) error {
		return nil
	})

	go b.Go()
	defer b.Stop()
	broker.invoke()

	rm := collect()
	want := []attribute.KeyValue{
		attribute.String("messaging.destination.name", "fake-broker"),
		attribute.String("ezbus.message_name", "FakeMessage"),
		attribute.String("ezbus.outcome", "ok"),
	}
	assert.Equal(t, int64(1), sumOf(rm, "messaging.client.consumed.messages", want...))
	count, attempts := histogramOf[int64](rm, "ezbus.process.attempts", want...)
	assert.Equal(t, uint64(1), count)
	assert.Equal(t, int64(1), attempts)
	count, _ = histogramOf[float64](rm, "messaging.process.duration", want...)
	assert.Equal(t, uint64(1), count)
}

func TestMessageOnErrorQueueIsCountedWithFiveAttempts(t *testing.T) {
	collect := collectMetrics(t)
	rtr.Handle("FakeMessage", func(m Message) error {
		return errors.New("no luck")
	})

	go b.Go()
	defer b.Stop()
	broker.invoke()

	rm := collect()
	want := []attribute.KeyValue{attribute.String("ezbus.outcome", "error_queue")}
	assert.Equal(t, int64(1), sumOf(rm, "messaging.client.consumed.messages", want...))
	_, attempts := histogramOf[int64](rm, "ezbus.process.attempts", want...)
	assert.Equal(t, int64(5), attempts)
}

func TestDiscardedMessageIsCounted(t *testing.T) {
	collect := collectMetrics(t)

	b.(*bus).handle(NewMessage(map[string]string{headers.MessageName: "NoSuchMessage"}, nil))

	rm := collect()
	assert.Equal(t, int64(1), sumOf(rm, "messaging.client.consumed.messages",
		attribute.String("ezbus.message_name", "NoSuchMessage"),
		attribute.String("ezbus.outcome", "discarded")))
}
