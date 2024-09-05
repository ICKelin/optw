package quic

import (
	"github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
	"time"
)

func TestQuic(t *testing.T) {
	convey.Convey("test quic transport", t, func() {
		convey.Convey("test quic normaly", func() {
			l := NewListener("127.0.0.1:3445")
			err := l.Listen()
			convey.So(err, convey.ShouldBeNil)
			defer l.Close()

			sndbuf := "test buffer"
			go func() {
				conn, err := l.Accept()
				if err != nil {
					t.Error("err should be nil")
				}
				defer conn.Close()

				stream, err := conn.AcceptStream()
				if err != nil {
					t.Error("err should be nil")
				}
				defer stream.Close()

				buf := make([]byte, len(sndbuf))
				_, err = io.ReadFull(stream, buf)
				if err != nil {
					t.Error("err should be nil")
				}
				_, err = stream.Write(buf)
				if err != nil {
					t.Error("err should be nil")
				}
			}()

			d := NewDialer("127.0.0.1:3445")
			conn, err := d.Dial()
			convey.So(err, convey.ShouldBeNil)
			defer conn.Close()

			stream, err := conn.OpenStream()
			convey.So(err, convey.ShouldBeNil)
			defer stream.Close()

			_, err = stream.Write([]byte(sndbuf))
			convey.So(err, convey.ShouldBeNil)

			buf := make([]byte, len(sndbuf))
			_, err = io.ReadFull(stream, buf)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(buf), convey.ShouldEqual, sndbuf)
		})

		convey.Convey("test reconnect", func() {
			l := NewListener("127.0.0.1:3445")
			err := l.Listen()
			convey.So(err, convey.ShouldBeNil)

			sndbuf := "test buffer"
			go func() {
				conn, err := l.Accept()
				if err != nil {
					t.Error("err should be nil")
				}
				defer conn.Close()

				stream, err := conn.AcceptStream()
				if err != nil {
					t.Error("err should be nil")
				}
				defer stream.Close()

				buf := make([]byte, len(sndbuf))
				_, err = io.ReadFull(stream, buf)
				if err != nil {
					t.Error("err should be nil")
				}
				_, err = stream.Write(buf)
				if err != nil {
					t.Error("err should be nil")
				}
			}()

			d := NewDialer("127.0.0.1:3445")
			conn, err := d.Dial()
			convey.So(err, convey.ShouldBeNil)
			defer conn.Close()

			stream, err := conn.OpenStream()
			convey.So(err, convey.ShouldBeNil)
			defer stream.Close()

			_, err = stream.Write([]byte(sndbuf))
			convey.So(err, convey.ShouldBeNil)

			buf := make([]byte, len(sndbuf))
			_, err = io.ReadFull(stream, buf)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(buf), convey.ShouldEqual, sndbuf)

			_, err = conn.OpenStream()
			t.Log(err)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(conn.IsClosed(), convey.ShouldEqual, false)
			conn.Close()
			convey.So(conn.IsClosed(), convey.ShouldBeTrue)
		})
	})
}

func TestQuicAuth(t *testing.T) {
	convey.Convey("test optw transport/quic auth", t, func() {
		convey.Convey("test auth success", func() {
			l := NewListener("127.0.0.1:2001")
			l.SetAuthFunc(func(token string) bool {
				if token == "test auth" {
					return true
				}
				return false
			})
			l.Listen()
			defer l.Close()
			d := NewDialer("127.0.0.1:2001")
			d.SetAccessToken("test auth")

			go func() {
				_, err := l.Accept()
				if err != nil {
					t.Error("err should be nil")
				}
			}()

			time.Sleep(time.Second * 1)
			conn, err := d.Dial()
			convey.So(err, convey.ShouldBeNil)
			defer conn.Close()
		})

		convey.Convey("test auth fail", func() {
			l := NewListener("127.0.0.1:2001")
			l.SetAuthFunc(func(token string) bool {
				if token == "test auth" {
					return true
				}
				return false
			})
			l.Listen()
			defer l.Close()
			d := NewDialer("127.0.0.1:2001")
			d.SetAccessToken("invalid test auth")

			go func() {
				_, err := l.Accept()
				if err == nil {
					t.Error("err should not be nil")
				}
			}()

			time.Sleep(time.Second * 1)
			_, err := d.Dial()
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("no auth test", func() {
			l := NewListener("127.0.0.1:2001")
			l.Listen()
			d := NewDialer("127.0.0.1:2001")

			go func() {
				_, err := l.Accept()
				if err != nil {
					t.Errorf("err should be nil, got %v", err)
				}
			}()

			time.Sleep(time.Second * 1)
			_, err := d.Dial()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
