package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptpclient "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/clientset/versioned"
	"github.com/zshi-redhat/kube-ptp-daemon/pkg/config"
	"github.com/zshi-redhat/kube-ptp-daemon/pkg/daemon"

        "k8s.io/client-go/kubernetes"
)

type cliParams struct {
	updateInterval	int
	logLevel	string
}

// Parse Command line flags
func flagInit(cp *cliParams) {
        flag.IntVar(&cp.updateInterval, "update-interval", config.DefaultUpdateInterval, "Interval to update PTP status")
        flag.StringVar(&cp.logLevel, "log-level", config.DefaultLogLevel, "Level of log message")
}


func main() {
	cp := &cliParams{}
	flag.Parse()
	flagInit(cp)

	config.SetLogLevel(cp.logLevel)

	logging.Debugf("log level set to: %s", cp.logLevel)
	logging.Debugf("resync period set to: %d [s]", cp.updateInterval)

	cfg, err := config.GetKubeConfig()
	if err != nil {
		logging.Errorf("get kubeconfig failed: %v", err)
		return
	}
	logging.Debugf("successfully get kubeconfig")

        kubeClient, err := kubernetes.NewForConfig(cfg)
        if err != nil {
                logging.Errorf("cannot create new config for kubeClient: %v", err)
                return
        }

	ptpClient, err := ptpclient.NewForConfig(cfg)
	if err != nil {
		logging.Errorf("cannot create new config for ptpClient: %v", err)
		return
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	err = daemon.New(
		os.Getenv("PTP_NODE_NAME"),
		daemon.PtpNamespace, ptpClient, kubeClient, stopCh,
	).Run()
	if err != nil {
		logging.Errorf("cannot run daemon: %v", err)
	}

	tickerPull := time.NewTicker(time.Second * time.Duration(cp.updateInterval))
	defer tickerPull.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-tickerPull.C:
			logging.Debugf("ticker pull")
		case sig := <-sigCh:
			logging.Debugf("signal received, shutting down", sig)
			return
		}
	}
}
