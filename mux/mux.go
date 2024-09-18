package mux

import (
	"fmt"
	"github.com/ICKelin/optw"
	"net"
	"time"

	"github.com/xtaci/smux"
)

var _ optw.Listener = &Listener{}
var _ optw.Dialer = &Dialer{}
var _ optw.Conn = &Conn{}

type Dialer struct {
	remote      string
	accessToken string
}

func (d *Dialer) SetAccessToken(accessToken string) {
	d.accessToken = accessToken
}

type Listener struct {
	laddr string
	net.Listener
	authFn func(token string) bool
}

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

func NewDialer(remote string) optw.Dialer {
	return &Dialer{remote: remote}
}

func (d *Dialer) Dial() (optw.Conn, error) {
	conn, err := net.Dial("tcp", d.remote)
	if err != nil {
		return nil, err
	}

	// enable auth
	if len(d.accessToken) > 0 {
		conn.SetDeadline(time.Now().Add(time.Second * 5))
		err = optw.AuthRequest(conn, d.accessToken)
		conn.SetDeadline(time.Time{})
		if err != nil {
			conn.Close()
			return nil, err
		}
	}

	cfg := smux.DefaultConfig()
	cfg.KeepAliveTimeout = time.Second * 40
	cfg.KeepAliveInterval = time.Second * 3
	mux, err := smux.Client(conn, cfg)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func NewListener(laddr string) *Listener {
	return &Listener{laddr: laddr}
}

func (l *Listener) Accept() (optw.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// enable auth
	if l.authFn != nil {
		conn.SetDeadline(time.Now().Add(time.Second * 5))
		err := optw.VerifyAuth(conn, l.authFn)
		conn.SetDeadline(time.Time{})
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("auth fail: %v", err)
		}
	}
	cfg := smux.DefaultConfig()
	cfg.KeepAliveTimeout = time.Second * 10
	cfg.KeepAliveInterval = time.Second * 3
	mux, err := smux.Server(conn, cfg)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func (l *Listener) Listen() error {
	listener, err := net.Listen("tcp", l.laddr)
	if err != nil {
		return err
	}

	l.Listener = listener
	return nil
}

func (l *Listener) SetAuthFunc(f func(token string) bool) {
	l.authFn = f
}
