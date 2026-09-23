package rabbitmq

// Option configures a Broker. Pass options to NewBroker.
type Option func(*config)

type config struct {
	url                string
	prefetchCount      int
	queueNameDelimiter string
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
