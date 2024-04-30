package quic

import (
	quic_go "github.com/quic-go/quic-go"
	"net"
)

type Stream struct {
	rawConn *Conn
	quic_go.Stream
}

func (s *Stream) RemoteAddr() net.Addr {
	return s.rawConn.RemoteAddr()
}

func (s *Stream) LocalAddr() net.Addr {
	return s.rawConn.LocalAddr()
}
