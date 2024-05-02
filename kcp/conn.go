package kcp

import (
	"github.com/ICKelin/optw"
	"net"
	"time"

	"github.com/xtaci/smux"
)

var _ optw.Conn = &Conn{}

type Conn struct {
	mux *smux.Session
}

func (c *Conn) OpenStream() (optw.Stream, error) {
	stream, err := c.mux.OpenStream()
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (c *Conn) AcceptStream() (optw.Stream, error) {
	return c.mux.AcceptStream()
}

func (c *Conn) Close() {
	c.mux.Close()
}

func (c *Conn) IsClosed() bool {
	return c.mux.IsClosed()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.mux.RemoteAddr()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.mux.LocalAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.mux.SetDeadline(t)
}
