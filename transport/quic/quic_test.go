package quic

import (
	"github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func TestQuic(t *testing.T) {
	convey.Convey("test quic transport", t, func() {
		convey.Convey("test quic normaly", func() {
			l := NewListener("127.0.0.1:3445")
			err := l.Listen()
			convey.So(err, convey.ShouldBeNil)

			sndbuf := make([]byte, 10)
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

				buf := make([]byte, 10)
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

			_, err = stream.Write(sndbuf)
			convey.So(err, convey.ShouldBeNil)

			buf := make([]byte, 10)
			_, err = io.ReadFull(stream, buf)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
