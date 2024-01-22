package kcp

import (
	"crypto/sha1"
	"encoding/json"
	"golang.org/x/crypto/pbkdf2"
	"net"

	"github.com/ICKelin/optw/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Listener = &Listener{}

type Listener struct {
	laddr  string
	config KCPConfig
	key    string
	*kcpgo.Listener
}

func NewListener(laddr string, key string, rawConfig json.RawMessage) *Listener {
	l := &Listener{}
	if len(rawConfig) <= 0 {
		l.config = defaultConfig
	} else {
		cfg := KCPConfig{}
		json.Unmarshal(rawConfig, &cfg)
		l.config = cfg
	}

	l.laddr = laddr
	l.key = key
	return l
}

func (l *Listener) Listen() error {
	cfg := l.config
	var block kcpgo.BlockCrypt
	pass := pbkdf2.Key([]byte(l.key), []byte(SALT), 4096, 32, sha1.New)

	switch cfg.Crypt {
	case "null":
		block = nil
	case "sm4":
		block, _ = kcpgo.NewSM4BlockCrypt(pass[:16])
	case "tea":
		block, _ = kcpgo.NewTEABlockCrypt(pass[:16])
	case "xor":
		block, _ = kcpgo.NewSimpleXORBlockCrypt(pass)
	case "none":
		block, _ = kcpgo.NewNoneBlockCrypt(pass)
	case "aes-128":
		block, _ = kcpgo.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		block, _ = kcpgo.NewAESBlockCrypt(pass[:24])
	case "blowfish":
		block, _ = kcpgo.NewBlowfishBlockCrypt(pass)
	case "twofish":
		block, _ = kcpgo.NewTwofishBlockCrypt(pass)
	case "cast5":
		block, _ = kcpgo.NewCast5BlockCrypt(pass[:16])
	case "3des":
		block, _ = kcpgo.NewTripleDESBlockCrypt(pass[:24])
	case "xtea":
		block, _ = kcpgo.NewXTEABlockCrypt(pass[:16])
	case "salsa20":
		block, _ = kcpgo.NewSalsa20BlockCrypt(pass)
	default:
		cfg.Crypt = "aes"
		block, _ = kcpgo.NewAESBlockCrypt(pass)
	}

	kcpLis, err := kcpgo.ListenWithOptions(l.laddr, block, 10, 3)
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
	kcpconn, err := l.Listener.AcceptKCP()
	if err != nil {
		return nil, err
	}

	kcpconn.SetStreamMode(true)
	kcpconn.SetWriteDelay(false)
	kcpconn.SetNoDelay(cfg.Nodelay, cfg.Interval, cfg.Resend, cfg.Nc)
	kcpconn.SetWindowSize(cfg.RcvWnd, cfg.SndWnd)
	kcpconn.SetMtu(cfg.Mtu)
	kcpconn.SetACKNoDelay(cfg.AckNoDelay)
	kcpconn.SetReadBuffer(cfg.Rcvbuf)
	kcpconn.SetWriteBuffer(cfg.SndBuf)
	mux, err := smux.Server(kcpconn, nil)
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
