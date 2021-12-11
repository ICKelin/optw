package target

import (
	"github.com/ICKelin/optw/internal/logs"
	"github.com/ICKelin/optw/transport"
	"io"
	"net"
	"sync"
)

type Forward struct {
	listener   transport.Listener
	targetAddr string
	mempool    sync.Pool
}

func NewForward(listener transport.Listener, targetAddr string) *Forward {
	return &Forward{
		listener:   listener,
		targetAddr: targetAddr,
		mempool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024*64)
			},
		},
	}
}

func (f *Forward) ListenAndServe() error {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			logs.Debug("accept fail: %v", err)
			break
		}

		logs.Debug("accept new connection: %v", conn.RemoteAddr())
		go f.forward(conn)
	}

	return nil
}

func (f *Forward) forward(conn transport.Conn) {
	defer conn.Close()

	for {
		stream, err := conn.AcceptStream()
		if err != nil {
			logs.Error("accept stream for %v fail; %v", conn.RemoteAddr(), err)
			break
		}

		go f.handleStream(stream)
	}
}

func (f *Forward) handleStream(stream transport.Stream) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// dial target address
	targetConn, err := net.Dial("tcp", f.targetAddr)
	if err != nil {
		logs.Error("dial target fail: %v", err)
		return
	}
	logs.Debug("open a new connection to target %v", f.targetAddr)

	go func() {
		defer wg.Done()
		defer targetConn.Close()
		defer stream.Close()
		obj := f.mempool.Get()
		defer f.mempool.Put(obj)
		buf := obj.([]byte)
		io.CopyBuffer(targetConn, stream, buf)
	}()

	go func() {
		defer wg.Done()
		defer targetConn.Close()
		defer stream.Close()
		obj := f.mempool.Get()
		defer f.mempool.Put(obj)
		buf := obj.([]byte)
		_, err := io.CopyBuffer(stream, targetConn, buf)
		logs.Debug("close copy conn->stream: %v", err)
	}()

	wg.Wait()
	logs.Debug("connection %v closed", targetConn.RemoteAddr())
}
