package rabbitmq

//Configurer for rabbitmq
type Configurer interface {
	URL(string)
	PrefetchCount(int)
	QueueNameDelimiter(string)
}

type config struct {
	url                string
	prefetchCount      int
	queueNameDelimiter string
}

func (c *config) URL(v string) {
	c.url = v
}

func (c *config) PrefetchCount(i int) {
	c.prefetchCount = i
}

func (c *config) QueueNameDelimiter(v string) {
	c.queueNameDelimiter = v
}
