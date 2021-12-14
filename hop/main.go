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

	// initial route table
	routeTable := NewRouteTable()
	for _, c := range routeCfg {
		for _, hopCfg := range c.HopConfig {
			go routeTable.Add(hopCfg.Scheme, hopCfg.HopAddr, hopCfg.RawConfig)
		}

		// initial hop
		lisCfg := c.ListenerConfig
		h := NewHop(lisCfg.Scheme, lisCfg.ListenAddr, routeTable)
		go h.Serve()
	}

	select {}
}
