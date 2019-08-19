package main

import (
	"flag"
	"os"
	"time"

	"github.com/golang/glog"
	ptputils "github.com/zshi-redhat/kube-ptp-daemon/pkg/utils"
)

const (
	defaultUpdateInterval = "5"
)

type cliParams struct {
	updateInterval	string
}

// Parse Command line flags
func flagInit(cp *cliParams) {
        flag.StringVar(&cp.updateInterval, "update-interval",
		defaultUpdateInterval, "Interval to update PTP status")
}

func main() {
	cp := &cliParams{}
	flag.Parse()
	flagInit(cp)

	glog.Infof("Resync period set to: %s [s]", cp.updateInterval)

	tickerPull := time.NewTicker(time.Second * time.Duration(cp.updateInterval))
	defer tickerPull.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-tickerPull.C:
		glog.Infof("ticker pull")
		nics, err := ptputils.DiscoverPTPDevices()
		if err != nil {
			glog.Infof("ticker pull failed")
		} else {
			glog.Infof("PTP capable NICs: %v", nics)
		}
	case sig := <-sigCh:
		glog.Infof("signal received, shutting down", sig)
		return
	}
}
