package transport_api

import (
	"errors"
	"github.com/ICKelin/optw/transport"
	"github.com/ICKelin/optw/transport/kcp"
	"github.com/ICKelin/optw/transport/mux"
	"github.com/ICKelin/optw/transport/quic"
)

const (
	protoKCP    = "kcp"
	protoTCPMux = "mux"
	protoQuic   = "quic"
)

var (
	errUnsupported = errors.New("transport_api: unsupported protocol")
)

func NewListen(scheme, addr, cfg string) (transport.Listener, error) {
	var listener transport.Listener
	switch scheme {
	case protoKCP:
		listener = kcp.NewListener(addr, []byte(cfg))
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

func NewDialer(scheme, addr, cfg string) (transport.Dialer, error) {
	var dialer transport.Dialer
	switch scheme {
	case protoKCP:
		dialer = kcp.NewDialer(addr, []byte(cfg))
	case protoTCPMux:
		dialer = mux.NewDialer(addr)
	case protoQuic:
		dialer = quic.NewDialer(addr)
	default:
		return nil, errUnsupported
	}

	return dialer, nil
}
