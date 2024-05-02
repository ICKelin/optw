package transport_api

import (
	"errors"
	"github.com/ICKelin/optw"
	kcp2 "github.com/ICKelin/optw/kcp"
	"github.com/ICKelin/optw/mux"
	"github.com/ICKelin/optw/quic"
)

const (
	protoKCP    = "kcp"
	protoTCPMux = "mux"
	protoQuic   = "quic"
)

var (
	errUnsupported = errors.New("transport_api: unsupported protocol")
)

func NewListen(scheme, addr, cfg string) (optw.Listener, error) {
	var listener optw.Listener
	switch scheme {
	case protoKCP:
		listener = kcp2.NewListener(addr, []byte(cfg))
	case protoTCPMux:
		listener = mux.NewListener(addr)
	case protoQuic:
		listener = quic.NewListener(addr)
	default:
		return nil, errUnsupported
	}

	err := listener.Listen()
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func NewDialer(scheme, addr, cfg string) (optw.Dialer, error) {
	var dialer optw.Dialer
	switch scheme {
	case protoKCP:
		dialer = kcp2.NewDialer(addr, []byte(cfg))
	case protoTCPMux:
		dialer = mux.NewDialer(addr)
	case protoQuic:
		dialer = quic.NewDialer(addr)
	default:
		return nil, errUnsupported
	}

	return dialer, nil
}
