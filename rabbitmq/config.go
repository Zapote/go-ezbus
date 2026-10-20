package rabbitmq

import "time"

// Option configures a Broker. Pass options to NewBroker.
type Option func(*config)

type config struct {
	url                string
	prefetchCount      int
	queueNameDelimiter string
	drainTimeout       time.Duration
	reconnectAttempts  int
	reconnectDelay     time.Duration
	exclusive          bool
}

// WithURL sets the AMQP URL to connect to.
// Default amqp://guest:guest@localhost:5672
func WithURL(url string) Option {
	return func(c *config) {
		c.url = url
	}
}

// WithPrefetchCount sets how many unacknowledged messages the broker
// takes at a time. Default 100.
func WithPrefetchCount(n int) Option {
	return func(c *config) {
		c.prefetchCount = n
	}
}

// WithQueueNameDelimiter sets what goes between the queue name and the
// error suffix, as in my-queue-error. Default "-".
func WithQueueNameDelimiter(d string) Option {
	return func(c *config) {
		c.queueNameDelimiter = d
	}
}

// WithDrainTimeout sets how long Stop waits for the handler that is
// running. Keep it below the time the process is given to exit, as in
// terminationGracePeriodSeconds. Default 20 seconds.
func WithDrainTimeout(d time.Duration) Option {
	return func(c *config) {
		c.drainTimeout = d
	}
}

// WithExclusiveQueue gives each instance a queue of its own, named queue@host,
// that lives as long as the connection. Every instance of a service then gets
// each message it subscribes to instead of sharing them with the others. The
// error queue lives as long. The broker declares no exchange and so cannot
// publish, and after a reconnect it declares the queue and binds its
// subscriptions again.
func WithExclusiveQueue() Option {
	return func(c *config) {
		c.exclusive = true
	}
}
