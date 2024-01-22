package target

import (
	"flag"
	"fmt"
	"github.com/ICKelin/optw/internal/logs"
	"github.com/ICKelin/optw/transport/transport_api"
)

func Main() {
	flgConf := flag.String("c", "", "config path")
	flag.Parse()

	cfg, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	logs.Init("target.log", "debug", 10)

	listener, err := transport_api.NewListen(cfg.ListenerConfig.Scheme,
		cfg.ListenerConfig.ListenAddr,
		cfg.ListenerConfig.Key,
		cfg.ListenerConfig.Cfg)

	if err != nil {
		logs.Error("new listener fail: %v", err)
		return
	}
	defer listener.Close()

	fw := NewForward(listener, cfg.TargetConfig.Address)
	logs.Error(fw.ListenAndServe())
}
