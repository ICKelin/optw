package quic

import (
	"context"
	"github.com/ICKelin/optw/transport"
	quic_go "github.com/quic-go/quic-go"
	"net"
	"time"
)

type Conn struct {
	close bool
	conn  quic_go.Connection
}

func (c *Conn) OpenStream() (transport.Stream, error) {
	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}

	return &Stream{rawConn: c, Stream: stream}, nil
}

func (c *Conn) AcceptStream() (transport.Stream, error) {
	stream, err := c.conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}

	return &Stream{rawConn: c, Stream: stream}, nil
}

func (c *Conn) Close() {
	c.conn.CloseWithError(0, "")
	c.close = true
}

func (c *Conn) IsClosed() bool {
	return c.close
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}
