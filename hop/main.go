package hop

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

	logs.Init("hop.log", "debug", 10)
	logs.Debug("hop config: %v", cfg)

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

	f := NewForward(lisCfg.Scheme, lisCfg.ListenAddr, routeTable)
	switch f.scheme {
	case "tcp":
		err = f.ServeTCP()
	default:
		err = f.ServeMux()
	}

	logs.Error("hop exist: %v", err)
}
