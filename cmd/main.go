package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	ptpclient "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/clientset/versioned"
	"github.com/zshi-redhat/kube-ptp-daemon/pkg/config"
	"github.com/zshi-redhat/kube-ptp-daemon/pkg/daemon"

        "k8s.io/client-go/kubernetes"
)

type cliParams struct {
	updateInterval	int
}

// Parse Command line flags
func flagInit(cp *cliParams) {
        flag.IntVar(&cp.updateInterval, "update-interval", config.DefaultUpdateInterval, "Interval to update PTP status")
}


func main() {
	cp := &cliParams{}
	flag.Parse()
	flagInit(cp)

	glog.Infof("resync period set to: %d [s]", cp.updateInterval)

	cfg, err := config.GetKubeConfig()
	if err != nil {
		glog.Errorf("get kubeconfig failed: %v", err)
		return
	}
	glog.Infof("successfully get kubeconfig")

        kubeClient, err := kubernetes.NewForConfig(cfg)
        if err != nil {
                glog.Errorf("cannot create new config for kubeClient: %v", err)
                return
        }

	ptpClient, err := ptpclient.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("cannot create new config for ptpClient: %v", err)
		return
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	err = daemon.New(
		os.Getenv("PTP_NODE_NAME"),
		daemon.PtpNamespace,
		ptpClient,
		kubeClient,
		stopCh,
	).Run()
	if err != nil {
		glog.Errorf("cannot run daemon: %v", err)
	}

	tickerPull := time.NewTicker(time.Second * time.Duration(cp.updateInterval))
	defer tickerPull.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-tickerPull.C:
			glog.Infof("ticker pull")
		case sig := <-sigCh:
			glog.Infof("signal received, shutting down", sig)
			return
		}
	}
}
