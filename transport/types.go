package transport

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// Dialer defines transport_api dialer for client side
type Dialer interface {
	Dial() (Conn, error)
	SetAccessToken(accessToken string)
}

// Listener defines transport_api listener for server side
type Listener interface {
	Listen() error
	// Accept returns a connection
	// if an error occurs, it may suit each implements error
	Accept() (Conn, error)

	// Close close a listener
	Close() error

	// Addr returns address of listener
	Addr() net.Addr

	SetAuthFunc(func(token string) bool)
}

// Conn defines a transport_api connection
type Conn interface {
	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
	Close()
	IsClosed() bool
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetDeadline(t time.Time) error
}

// Stream defines a transport_api stream base on
// Conn.OpenStream or Conn.AcceptStream
type Stream interface {
	Write(buf []byte) (int, error)
	Read(buf []byte) (int, error)
	Close() error
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetDeadline(t time.Time) error
}

func VerifyAuth(conn io.Reader, authFn func(token string) bool) error {
	hdr := make([]byte, 2)
	_, err := io.ReadFull(conn, hdr)
	if err != nil {
		return fmt.Errorf("read auth hdr fail: %v", err)
	}
	tokenLen := binary.BigEndian.Uint16(hdr)
	token := make([]byte, tokenLen)
	_, err = io.ReadFull(conn, token)
	if err != nil {
		return fmt.Errorf("read access token fail: %v", err)
	}

	ok := authFn(string(token))
	if !ok {
		return fmt.Errorf("verify token %s fail", token)
	}
	return nil
}
