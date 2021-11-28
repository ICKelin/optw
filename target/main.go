package target

import (
	"flag"
	"github.com/ICKelin/gtun/transport/transport_api"
	"github.com/ICKelin/optw/internal/logs"
)

func Main() {
	flgScheme := flag.String("scheme", "", "listen scheme(mux/kcp)")
	flgListenAddr := flag.String("l", "", "listen address")
	flgTargetAddr := flag.String("t", "", "target address")
	flag.Parse()

	logs.Init("target.log", "debug", 10)

	listener, err := transport_api.NewListen(*flgScheme, *flgListenAddr, "")
	if err != nil {
		logs.Error("new listener fail: %v", err)
		return
	}
	defer listener.Close()

	fw := NewForward(listener, *flgTargetAddr)
	logs.Error(fw.ListenAndServe())
}
