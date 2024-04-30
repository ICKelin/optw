package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/ICKelin/optw/transport"
	quic_go "github.com/quic-go/quic-go"
	"math/big"
	"net"
	"time"
)

var _ transport.Listener = &Listener{}
var _ transport.Dialer = &Dialer{}

var nextProtocols = []string{
	"ickelin/optw",
}

type Listener struct {
	addr     string
	listener *quic_go.Listener
	authFn   func(token string) bool
}

func NewListener(addr string) *Listener {
	return &Listener{addr: addr}
}

func (l *Listener) Listen() error {
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		return err
	}
	listener, err := quic_go.ListenAddr(l.addr, tlsConfig, &quic_go.Config{
		KeepAlivePeriod: time.Second * 10,
	})
	if err != nil {
		return err
	}
	l.listener = listener
	return nil
}

func (l *Listener) Accept() (transport.Conn, error) {
	conn, err := l.listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{close: false, conn: conn}, nil
}

func (l *Listener) Close() error {
	if l.listener != nil {
		l.listener.Close()
	}
	return nil
}

func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *Listener) SetAuthFunc(f func(token string) bool) {
	l.authFn = f
}

type Dialer struct {
	addr        string
	accessToken string
}

func NewDialer(addr string) *Dialer {
	return &Dialer{addr: addr}
}

func (d *Dialer) Dial() (transport.Conn, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         nextProtocols,
	}
	conn, err := quic_go.DialAddr(context.Background(), d.addr, tlsConf, &quic_go.Config{
		KeepAlivePeriod: time.Second * 10,
	})
	if err != nil {
		return nil, err
	}

	return &Conn{close: false, conn: conn}, nil
}

func (d *Dialer) SetAccessToken(accessToken string) {
	d.accessToken = accessToken
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   nextProtocols,
	}, nil
}
