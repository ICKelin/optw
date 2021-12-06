package access

import (
	"flag"
	"fmt"
	"github.com/ICKelin/optw/internal/logs"
)

func Main() {
	flgConf := flag.String("c", "", "config file path")
	flag.Parse()

	cfg, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	logs.Init("access.log", "debug", 10)
	logs.Debug("access config: %v", cfg)

	routeCfg := cfg.RouteConfig
	// initial local listener
	lisCfg := routeCfg.ListenerConfig

	routeTable := NewRouteTable()

	// initial next hop dialer
	for _, nexthop := range routeCfg.NexthopConfig {
		err = routeTable.Add(nexthop.Scheme, nexthop.NexthopAddr, nexthop.RawConfig)
		if err != nil {
			logs.Error("add route table fail: %v", err)
			continue
		}
	}

	f := NewForward(lisCfg.ListenAddr, routeTable)

	if err := f.ServeTCP(); err != nil {
		logs.Error("access exist: %v", err)
	}

	select {}
}
