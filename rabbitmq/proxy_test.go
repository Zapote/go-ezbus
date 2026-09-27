package rabbitmq

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

// proxy sits between the broker under test and RabbitMQ, so that a test can
// cut the connection the way a network failure does, or silence the server
// the way a hung one does.
type proxy struct {
	target   string
	accepted atomic.Int32
	muted    atomic.Bool

	mu    sync.Mutex
	addr  string
	ln    net.Listener
	conns []net.Conn
}

func newProxy(t *testing.T, target string) *proxy {
	t.Helper()
	p := &proxy{target: target, addr: "127.0.0.1:0"}
	p.listen(t)
	t.Cleanup(p.stop)
	return p
}

func (p *proxy) url() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return "amqp://guest:guest@" + p.addr
}

// listen accepts connections, on the same address after a stop.
func (p *proxy) listen(t *testing.T) {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()

	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		t.Fatalf("Listen: %s", err)
	}
	p.ln = ln
	p.addr = ln.Addr().String()

	go func() {
		for {
			client, err := ln.Accept()
			if err != nil {
				return
			}
			p.accepted.Add(1)
			p.forward(client)
		}
	}()
}

func (p *proxy) forward(client net.Conn) {
	server, err := net.Dial("tcp", p.target)
	if err != nil {
		client.Close()
		return
	}

	p.mu.Lock()
	p.conns = append(p.conns, client, server)
	p.mu.Unlock()

	go func() {
		io.Copy(server, client)
		server.Close()
	}()
	go func() {
		io.Copy(client, p.unlessMuted(server))
		client.Close()
	}()
}

// mute keeps the connections open but drops what the server sends.
func (p *proxy) mute() {
	p.muted.Store(true)
}

func (p *proxy) unlessMuted(server io.Reader) io.Reader {
	return readerFunc(func(buf []byte) (int, error) {
		for {
			n, err := server.Read(buf)
			if !p.muted.Load() || err != nil {
				return n, err
			}
		}
	})
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(buf []byte) (int, error) { return f(buf) }

// stop refuses new connections and cuts the open ones.
func (p *proxy) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ln != nil {
		p.ln.Close()
		p.ln = nil
	}
	for _, c := range p.conns {
		c.Close()
	}
	p.conns = nil
}
