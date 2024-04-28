package kcp

import (
	"github.com/smartystreets/goconvey/convey"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestKCP(t *testing.T) {
	convey.Convey("test optw transport/mux", t, func() {
		convey.Convey("test auth success", func() {
			l := NewListener("127.0.0.1:2001", nil)
			l.SetAuthFunc(func(token string) bool {
				if token == "test auth" {
					return true
				}
				return false
			})
			l.Listen()
			defer l.Close()
			d := NewDialer("127.0.0.1:2001", nil)
			d.SetAccessToken("test auth")

			go func() {
				_, err := l.Accept()
				if err != nil {
					t.Error("err should be nil, got ", err)
				}
			}()

			time.Sleep(time.Second * 1)
			conn, err := d.Dial()
			convey.So(err, convey.ShouldBeNil)
			defer conn.Close()
		})

		convey.Convey("test auth fail", func() {
			l := NewListener("127.0.0.1:2001", nil)
			l.SetAuthFunc(func(token string) bool {
				if token == "test auth" {
					return true
				}
				return false
			})
			l.Listen()
			defer l.Close()
			d := NewDialer("127.0.0.1:2001", nil)
			d.SetAccessToken("invalid test auth")

			go func() {
				_, err := l.Accept()
				if err == nil {
					t.Error("err should not be nil")
				}
				exist := strings.Contains(err.Error(), "auth fail")
				if !exist {
					t.Error("err should contains auth keyword")
				}
			}()

			time.Sleep(time.Second * 1)
			_, err := d.Dial()
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("no auth test", func() {
			l := NewListener("127.0.0.1:2001", nil)
			l.Listen()
			defer l.Close()
			d := NewDialer("127.0.0.1:2001", nil)

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := l.Accept()
				if err != nil {
					t.Error("err should be nil")
				}
			}()

			time.Sleep(time.Second * 1)
			conn, err := d.Dial()
			convey.So(err, convey.ShouldBeNil)
			_, err = conn.OpenStream()
			convey.So(err, convey.ShouldBeNil)
			wg.Wait()
			defer conn.Close()
		})
	})
}
