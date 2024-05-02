package main

import (
	"flag"
	"fmt"
	"github.com/ICKelin/optw"
	"github.com/ICKelin/optw/kcp"
	"github.com/ICKelin/optw/mux"
	"github.com/ICKelin/optw/quic"
	"io"
	"os"
	"time"
)

func main() {
	var protocol string
	var addr string
	flag.StringVar(&protocol, "protocol", "mux", "protocol(mux/kcp/quic)")
	flag.StringVar(&addr, "addr", "127.0.0.1:5002", "address for listener/dial")
	flag.Parse()

	var listener optw.Listener
	var dialer optw.Dialer
	switch protocol {
	case "quic":
		listener = quic.NewListener(addr)
		dialer = quic.NewDialer(addr)
	case "mux":
		listener = mux.NewListener(addr)
		dialer = mux.NewDialer(addr)
	case "kcp":
		listener = kcp.NewListener(addr, nil)
		dialer = kcp.NewDialer(addr, nil)
	default:
		panic("unsupported protocol")
	}

	listener.Listen()
	go func() {
		conn, err := dialer.Dial()
		if err != nil {
			fmt.Println("dialer dial fail: ", err)
			return
		}
		defer conn.Close()

		stream, err := conn.OpenStream()
		if err != nil {
			fmt.Println("dialer open stream fail: %v", err)
			return
		}
		defer stream.Close()

		tick := time.NewTicker(time.Second * 1)
		defer tick.Stop()
		i := 0
		for range tick.C {
			_, err = stream.Write([]byte(fmt.Sprintf("ping %d\n", i)))
			if err != nil {
				fmt.Println("dialer write fail: ", err)
				break
			}
			i += 1
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		go handleConn(conn)
	}

}

func handleConn(conn optw.Conn) {
	defer conn.Close()
	for {
		stream, err := conn.AcceptStream()
		if err != nil {
			fmt.Println("accept stream fail: ", err)
			break
		}
		go handleStream(stream)
	}
}

func handleStream(stream optw.Stream) {
	defer stream.Close()
	io.Copy(os.Stdout, stream)
}
