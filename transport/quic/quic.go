package quic

import (
	"github.com/ICKelin/optw/transport"
)

//var _ transport.Listener = &Listener{}
var _ transport.Dialer = &Dialer{}

//var _ transport.Conn = &Conn{}

type Dialer struct {
	remote string
}

func (d Dialer) Dial() (transport.Conn, error) {
	panic("implement me")
}
