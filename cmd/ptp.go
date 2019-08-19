package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptputils "github.com/zshi-redhat/kube-ptp-daemon/pkg/utils"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

func getConfig() (*rest.Config, error) {
        configFromFlags := func(kubeConfig string) (*rest.Config, error) {
                if _, err := os.Stat(kubeConfig); err != nil {
                        return nil, fmt.Errorf("Cannot stat kubeconfig '%s'", kubeConfig)
                }
                return clientcmd.BuildConfigFromFlags("", kubeConfig)
        }

        // If an env variable is specified with the config location, use that
        kubeConfig := os.Getenv("KUBECONFIG")
        if len(kubeConfig) > 0 {
                return configFromFlags(kubeConfig)
        }
        // If no explicit location, try the in-cluster config
        if c, err := rest.InClusterConfig(); err == nil {
                return c, nil
        }
        // If no in-cluster config, try the default location in the user's home directory
        if usr, err := user.Current(); err == nil {
                kubeConfig := filepath.Join(usr.HomeDir, ".kube", "config")
                return configFromFlags(kubeConfig)
        }

        return nil, fmt.Errorf("Could not locate a kubeconfig")
}

func main() {
	cp := &cliParams{}
	flag.Parse()
	flagInit(cp)

	if cp.logLevel != ""{
		logging.SetLogLevel(cp.logLevel)
	}

	logging.Debugf("Resync period set to: %d [s]", cp.updateInterval)

	config, err := getConfig()
	if err != nil {
		logging.Errorf("get kubeconfig failed: %v", err)
		return
	}
	logging.Debugf("kubeconfig: %v", config)

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
