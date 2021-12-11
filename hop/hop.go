package hop

import (
	"github.com/ICKelin/optw/internal/logs"
	"github.com/ICKelin/optw/transport"
	"github.com/ICKelin/optw/transport/transport_api"
	"io"
	"net"
	"sync"
)

type Hop struct {
	scheme     string
	addr       string
	routeTable *RouteTable
	mempool    sync.Pool
}

func NewHop(scheme, addr string, routeTable *RouteTable) *Hop {
	return &Hop{
		scheme:     scheme,
		addr:       addr,
		routeTable: routeTable,
		mempool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024*4)
			},
		},
	}
}

func (h *Hop) Serve() error {
	switch h.scheme {
	case "tcp":
		return h.ServeTCP()
	default:
		return h.ServeMux()
	}
}

func (h *Hop) ServeTCP() error {
	listener, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Debug("accept fail: %v", err)
			break
		}

		logs.Debug("accept new connection: %v", conn.RemoteAddr())
		go h.forward(conn)
	}

	return nil
}

func (h *Hop) ServeMux() error {
	listener, err := transport_api.NewListen(h.scheme, h.addr, "")
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Error("accept fail: %v", err)
			break
		}

		go h.handleMuxConn(conn)
	}
	return nil
}

func (h *Hop) handleMuxConn(conn transport.Conn) {
	defer conn.Close()

	for {
		stream, err := conn.AcceptStream()
		if err != nil {
			logs.Error("accept stream fail: %v", err)
			break
		}

		go func(stream transport.Stream) {
			logs.Warn("stream %s closed", stream.RemoteAddr())
			h.forward(stream)
		}(stream)
	}
}

func (h *Hop) forward(conn io.ReadWriteCloser) {
	entry, err := h.routeTable.Route()
	if err != nil {
		logs.Error("route fail: %v", err)
		return
	}
	logs.Debug("next hop:%v", entry.conn.RemoteAddr())

	stream, err := entry.conn.OpenStream()
	if err != nil {
		logs.Error("open next hop stream fail: %v", err)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer conn.Close()
		defer stream.Close()
		defer wg.Done()
		obj := h.mempool.Get()
		defer h.mempool.Put(obj)
		buf := obj.([]byte)
		io.CopyBuffer(conn, stream, buf)
	}()

	go func() {
		defer conn.Close()
		defer stream.Close()
		defer wg.Done()
		obj := h.mempool.Get()
		defer h.mempool.Put(obj)
		buf := obj.([]byte)
		_, err := io.CopyBuffer(stream, conn, buf)
		logs.Debug("close copy conn->stream: %v", err)
	}()

	wg.Wait()
}
