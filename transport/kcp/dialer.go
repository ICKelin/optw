package kcp

import (
	"crypto/sha1"
	"encoding/json"
	"golang.org/x/crypto/pbkdf2"

	"github.com/ICKelin/optw/transport"
	kcpgo "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

var _ transport.Dialer = &Dialer{}
var SALT = "opww"

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
	remote string
	key    string
	config KCPConfig
	block  kcpgo.BlockCrypt
}

func NewDialer(remote, key string, rawConfig json.RawMessage) *Dialer {
	dialer := &Dialer{remote: remote}
	if len(rawConfig) <= 0 {
		dialer.config = defaultConfig
	} else {
		cfg := KCPConfig{}
		json.Unmarshal(rawConfig, &cfg)
		dialer.config = cfg
	}
	dialer.key = key
	return dialer
}

func (dialer *Dialer) Dial() (transport.Conn, error) {
	cfg := dialer.config

	var block kcpgo.BlockCrypt
	pass := pbkdf2.Key([]byte(dialer.key), []byte(SALT), 4096, 32, sha1.New)

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

	kcpconn, err := kcpgo.DialWithOptions(dialer.remote, block, cfg.FecDataShards, cfg.FecParityShards)
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

	sess, err := smux.Client(kcpconn, nil)
	if err != nil {
		return nil, err
	}
	return &Conn{sess}, err
}
