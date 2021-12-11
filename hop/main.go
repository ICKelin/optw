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

	// initial route table
	// TODO: get route table from registry
	routeTable := NewRouteTable()
	for _, hopCfg := range routeCfg.HopConfig {
		err = routeTable.Add(hopCfg.Scheme, hopCfg.HopAddr, hopCfg.RawConfig)
		if err != nil {
			logs.Error("add route table fail: %v", err)
			continue
		}
	}

	// initial hop
	h := NewHop(lisCfg.Scheme, lisCfg.ListenAddr, routeTable)
	err = h.Serve()
	logs.Error("hop exist: %v", err)
}
