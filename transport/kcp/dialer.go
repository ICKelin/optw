package kcp

import (
	"encoding/binary"
	"encoding/json"
	"github.com/ICKelin/optw/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Dialer = &Dialer{}

var defaultConfig = KCPConfig{
	FecDataShards:   10,
	FecParityShards: 3,
	Nodelay:         1,
	Interval:        10,
	Resend:          2,
	Nc:              1,
	SndWnd:          1024,
	RcvWnd:          1024,
	Mtu:             1350,
	AckNoDelay:      true,
	Rcvbuf:          4194304,
	SndBuf:          4194304,
}

type Dialer struct {
	remote      string
	config      KCPConfig
	accessToken string
}

func (dialer *Dialer) SetAccessToken(accessToken string) {
	dialer.accessToken = accessToken
}

func NewDialer(remote string, rawConfig json.RawMessage) *Dialer {
	dialer := &Dialer{remote: remote}
	if len(rawConfig) <= 0 {
		dialer.config = defaultConfig
	} else {
		cfg := KCPConfig{}
		json.Unmarshal(rawConfig, &cfg)
		dialer.config = cfg
	}
	return dialer
}

func (dialer *Dialer) Dial() (transport.Conn, error) {
	cfg := dialer.config
	conn, err := kcpgo.DialWithOptions(dialer.remote, nil, cfg.FecDataShards, cfg.FecParityShards)
	if err != nil {
		return nil, err
	}

	// enable auth
	if len(dialer.accessToken) > 0 {
		hdr := make([]byte, 2)
		binary.BigEndian.PutUint16(hdr, uint16(len(dialer.accessToken)))
		_, err = conn.Write(append(hdr, []byte(dialer.accessToken)...))
		if err != nil {
			return nil, err
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

	sess, err := smux.Client(conn, nil)
	if err != nil {
		return nil, err
	}
	return &Conn{sess}, err
}
