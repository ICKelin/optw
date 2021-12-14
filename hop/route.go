package hop

import (
	"fmt"
	"github.com/ICKelin/optw/internal/logs"
	"github.com/ICKelin/optw/transport"
	"github.com/ICKelin/optw/transport/transport_api"
	"math"
	"sort"
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
	srtt              int64
	hitCount          int64
	conn              transport.Conn
}

type EntryList []*RouteEntry

func (l EntryList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l EntryList) Less(i, j int) bool {
	return l[i].hitCount < l[j].hitCount
}

func (l EntryList) Len() int {
	return len(l)
}

type RouteTable struct {
	tableMu  sync.RWMutex
	tableIdx int32
	table    EntryList
}

func NewRouteTable() *RouteTable {
	rt := &RouteTable{
		table: make(EntryList, 0),
	}

	go rt.healthCheck()
	return rt
}

func (r *RouteTable) healthCheck() {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()

	for range tick.C {
		aliveConn := make(EntryList, 0)

		r.tableMu.RLock()
		tb := r.table
		r.tableMu.RUnlock()
		for _, entry := range tb {
			if entry.conn.IsClosed() {
				logs.Error("next hop %s://%s disconnect", entry.scheme, entry.addr)
				go r.reconnect(entry)
			} else {
				aliveConn = append(aliveConn, entry)
				logs.Info("hop %s://%s hit count %d",
					entry.scheme, entry.addr, atomic.LoadInt64(&entry.hitCount))
			}
		}
		sort.Sort(aliveConn)

		r.tableMu.Lock()
		r.table = aliveConn
		r.tableMu.Unlock()
	}
}

func (r *RouteTable) newConn(scheme, addr, cfg string) (transport.Conn, error) {
	dialer, err := transport_api.NewDialer(scheme, addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("new dialer fail: %v", err)
	}

	conn, err := dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("dial fail: %v", err)
	}

	return conn, nil
}

func (r *RouteTable) newEntry(scheme, addr, cfg string) (*RouteEntry, error) {
	for {
		conn, err := r.newConn(scheme, addr, cfg)
		if err != nil {
			logs.Error("new conn fail: %v", err)
			time.Sleep(time.Second * 1)
			continue
		}

		entry := &RouteEntry{
			scheme: scheme,
			addr:   addr,
			cfg:    cfg,
			conn:   conn,
		}

		return entry, nil
	}
}

func (r *RouteTable) reconnect(e *RouteEntry) {
	var err error
	for {
		e, err = r.newEntry(e.scheme, e.addr, e.cfg)
		if err != nil {
			logs.Error("reconnect %s://%s fail, retrying")
			continue
		}

		logs.Info("reconnect %s://%s success", e.scheme, e.addr)
		r.table = append(r.table, e)
		break
	}
}

func (r *RouteTable) Add(scheme, addr, cfg string) error {
	entry, err := r.newEntry(scheme, addr, cfg)
	if err != nil {
		return err
	}

	r.tableMu.Lock()
	defer r.tableMu.Unlock()
	r.table = append(r.table, entry)
	logs.Debug("add route table:  %+v", entry)
	return nil
}

func (r *RouteTable) Del(scheme, addr string) {
	r.tableMu.Lock()
	defer r.tableMu.Unlock()
	idx := -1
	for i, entry := range r.table {
		if entry.scheme == scheme &&
			entry.addr == addr {
			idx = i
			entry.conn.Close()
			break
		}
	}

	if idx == -1 {
		return
	}

	// first element
	if idx == 0 {
		if len(r.table) == 0 {
			r.table = make(EntryList, 0)
		} else {
			r.table = r.table[idx+1:]
		}
		return
	}

	// last element
	if idx == len(r.table)-1 {
		r.table = r.table[:idx]
		return
	}

	r.table = append(r.table[:idx], r.table[idx+1:]...)
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
