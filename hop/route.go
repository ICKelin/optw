package hop

import (
	"fmt"
	"github.com/ICKelin/gtun/transport"
	"github.com/ICKelin/gtun/transport/transport_api"
	"github.com/ICKelin/optw/internal/logs"
	"math"
	"sync"
	"time"
)

var (
	errNoRoute = fmt.Errorf("no route to host")
	maxRtt     = math.MinInt32
)

type RouteEntry struct {
	scheme, addr, cfg string
	rtt               int32
	conn              transport.Conn
}

type RouteTable struct {
	// key: scheme://addr
	tableMu   sync.RWMutex
	table     map[string]*RouteEntry
	minRttKey string
}

func NewRouteTable() *RouteTable {
	rt := &RouteTable{
		table: make(map[string]*RouteEntry),
	}

	go rt.healthy()
	return rt
}

func (r *RouteTable) healthy() {
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
			}
		}
		r.table = aliveConn
		r.tableMu.Unlock()

		if len(deadConn) > 0 {
			go func(conns map[string]*RouteEntry) {
				for entryKey, entry := range conns {
					e, err := r.newEntry(entry.scheme, entry.addr, entry.cfg)
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

		if len(aliveConnForRtt) > 0 {
			minRtt := int64(-1)
			minRttKey := ""
			for entryKey, entry := range aliveConnForRtt {
				beg := time.Now()
				s, err := entry.conn.OpenStream()
				if err != nil {
					logs.Error("open stream to %s for rtt measurement fail: %v", entry.addr, err)
					continue
				}

				diff := time.Now().Sub(beg).Microseconds()
				logs.Info("next hop %s rtt cost %d micro seconds", entry.conn.RemoteAddr(), diff)
				if minRtt < diff {
					minRtt = diff
					minRttKey = entryKey
				}
				s.Close()
			}
			r.minRttKey = minRttKey
		}
	}
}

func (r *RouteTable) newEntry(scheme, addr, cfg string) (*RouteEntry, error) {
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
			scheme: scheme,
			addr:   addr,
			cfg:    cfg,
			conn:   conn,
		}
		return entry, nil
	}
}

func (r *RouteTable) Add(scheme, addr, cfg string) error {
	entry, err := r.newEntry(scheme, addr, cfg)
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
			break
		}
	}
}

func (r *RouteTable) Route() (*RouteEntry, error) {
	r.tableMu.RLock()
	defer r.tableMu.RUnlock()
	if len(r.table) <= 0 {
		return nil, errNoRoute
	}

	entry, ok := r.table[r.minRttKey]
	if !ok {
		logs.Info("minRtt value not exist, use random item")
		for _, e := range r.table {
			entry = e
			break
		}
	} else {
		logs.Debug("hit minRtt %s, hop %s", r.minRttKey, entry.addr)
	}

	return entry, nil
}
