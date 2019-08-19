package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptputils "github.com/zshi-redhat/kube-ptp-daemon/pkg/utils"
)

const (
	defaultUpdateInterval = 5
	defaultLogLevel       = "debug"
)

type cliParams struct {
	updateInterval	int
	logLevel	string
}

// Parse Command line flags
func flagInit(cp *cliParams) {
        flag.IntVar(&cp.updateInterval, "update-interval", defaultUpdateInterval, "Interval to update PTP status")
        flag.StringVar(&cp.logLevel, "log-level", defaultLogLevel, "Level of log message")
}

func main() {
	cp := &cliParams{}
	flag.Parse()
	flagInit(cp)

	if cp.logLevel != ""{
		logging.SetLogLevel(cp.logLevel)
	}

	logging.Debugf("Resync period set to: %d [s]", cp.updateInterval)

	nics, err := ptputils.DiscoverPTPDevices()
	if err != nil {
		logging.Debugf("discover PTP device failed: %v", err)
		return
	}
	logging.Debugf("PTP capable NICs: %v", nics)

	tickerPull := time.NewTicker(time.Second * time.Duration(cp.updateInterval))
	defer tickerPull.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-tickerPull.C:
		logging.Debugf("ticker pull")
	case sig := <-sigCh:
		logging.Debugf("signal received, shutting down", sig)
		return
	}
}
