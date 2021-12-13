package hop

import (
	"fmt"
	"github.com/ICKelin/optw/internal/logs"
	"github.com/ICKelin/optw/transport"
	"github.com/ICKelin/optw/transport/transport_api"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var (
	errNoRoute = fmt.Errorf("no route to host")
	maxRtt     = math.MinInt32
)

type RouteEntry struct {
	scheme, addr, cfg string
	rtt               int64
	loss              int64
	hitCount          int64
	conn              transport.Conn
	probeAddr         string
}

type RouteTable struct {
	// key: scheme://addr
	tableMu sync.RWMutex
	table   map[string]*RouteEntry
}

func NewRouteTable() *RouteTable {
	rt := &RouteTable{
		table: make(map[string]*RouteEntry),
	}

	go rt.healthCheck()
	return rt
}

func (r *RouteTable) healthCheck() {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()

	for range tick.C {
		deadConn := make(map[string]*RouteEntry)
		aliveConn := make(map[string]*RouteEntry)
		aliveConnForRtt := make(map[string]*RouteEntry, 0)

		r.tableMu.Lock()
		for entryKey, entry := range r.table {
			if entry.conn.IsClosed() {
				logs.Error("next hop %s disconnect", entryKey)
				deadConn[entryKey] = entry
			} else {
				aliveConn[entryKey] = entry
				aliveConnForRtt[entryKey] = entry
				logs.Info("hop %s hit count %d", entryKey, atomic.LoadInt64(&entry.hitCount))
			}
		}
		r.table = aliveConn
		r.tableMu.Unlock()

		if len(deadConn) > 0 {
			go func(conns map[string]*RouteEntry) {
				for entryKey, entry := range conns {
					e, err := r.newEntry(entry.scheme, entry.addr, entry.probeAddr, entry.cfg)
					if err != nil {
						logs.Debug("new entry fail: %v", err)
						continue
					}

					logs.Info("reconnect next hop %s", entryKey)

					r.tableMu.Lock()
					r.table[entryKey] = e
					r.tableMu.Unlock()
				}
			}(deadConn)
		}
	}
}

func (r *RouteTable) probe(scheme, probeAddr string) {}

func (r *RouteTable) newEntry(scheme, addr, probeAddr, cfg string) (*RouteEntry, error) {
	for {
		dialer, err := transport_api.NewDialer(scheme, addr, cfg)
		if err != nil {
			logs.Error("new dialer fail: %v", err)
			time.Sleep(time.Second * 1)
			continue
		}

		conn, err := dialer.Dial()
		if err != nil {
			logs.Error("dial fail: %v", err)
			time.Sleep(time.Second * 1)
			continue
		}

		entry := &RouteEntry{
			scheme:    scheme,
			addr:      addr,
			cfg:       cfg,
			conn:      conn,
			probeAddr: probeAddr,
		}

		go r.probe(scheme, probeAddr)
		return entry, nil
	}
}

func (r *RouteTable) Add(scheme, addr, probeAddr, cfg string) error {
	entry, err := r.newEntry(scheme, addr, probeAddr, cfg)
	if err != nil {
		return err
	}

	entryKey := fmt.Sprintf("%s://%s", scheme, addr)
	r.tableMu.Lock()
	defer r.tableMu.Unlock()
	r.table[entryKey] = entry
	logs.Debug("add route table: %s %+v", entryKey, entry)
	return nil
}

func (r *RouteTable) Del(scheme, addr string) {
	r.tableMu.Lock()
	defer r.tableMu.Unlock()
	for key, entry := range r.table {
		if entry.scheme == scheme &&
			entry.addr == addr {
			delete(r.table, key)
			entry.conn.Close()
			break
		}
	}
}

func (r *RouteTable) Route() (transport.Stream, error) {
	r.tableMu.RLock()
	defer r.tableMu.RUnlock()
	if len(r.table) <= 0 {
		return nil, errNoRoute
	}
	for _, e := range r.table {
		stream, err := e.conn.OpenStream()
		if err != nil {
			logs.Error("entry %s open stream fail: %v", e.conn.RemoteAddr(), err)
			continue
		}

		atomic.AddInt64(&e.hitCount, 1)
		return stream, nil
	}

	return nil, errNoRoute
}
