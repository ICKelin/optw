package kcp

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ICKelin/optw/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Listener = &Listener{}

type Listener struct {
	laddr  string
	config KCPConfig
	*kcpgo.Listener
	authFn func(token string) bool
}

func (l *Listener) SetAuthFunc(f func(token string) bool) {
	l.authFn = f
}

func NewListener(laddr string, rawConfig json.RawMessage) *Listener {
	l := &Listener{}
	if len(rawConfig) <= 0 {
		l.config = defaultConfig
	} else {
		cfg := KCPConfig{}
		json.Unmarshal(rawConfig, &cfg)
		l.config = cfg
	}

	l.laddr = laddr
	return l
}

func (l *Listener) Listen() error {
	kcpLis, err := kcpgo.ListenWithOptions(l.laddr, nil, 10, 3)
	if err != nil {
		return err
	}
	kcpLis.SetReadBuffer(4194304)
	kcpLis.SetWriteBuffer(4194304)
	l.Listener = kcpLis
	return nil
}

func (l *Listener) Accept() (transport.Conn, error) {
	cfg := l.config
	conn, err := l.Listener.AcceptKCP()
	if err != nil {
		return nil, err
	}

	if l.authFn != nil {
		err := transport.VerifyAuth(conn, l.authFn)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("auth fail: %v", err)
		}
	}

	conn.SetStreamMode(true)
	conn.SetWriteDelay(false)
	conn.SetNoDelay(cfg.Nodelay, cfg.Interval, cfg.Resend, cfg.Nc)
	conn.SetWindowSize(cfg.RcvWnd, cfg.SndWnd)
	conn.SetMtu(cfg.Mtu)
	conn.SetACKNoDelay(cfg.AckNoDelay)
	conn.SetReadBuffer(cfg.Rcvbuf)
	conn.SetWriteBuffer(cfg.SndBuf)
	mux, err := smux.Server(conn, nil)
	if err != nil {
		return nil, err
	}

	return &Conn{mux: mux}, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}
