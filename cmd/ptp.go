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
        "k8s.io/apimachinery/pkg/api/errors"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
        "k8s.io/client-go/tools/clientcmd"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
	ptpclient "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/clientset/versioned"
	ptpinformer "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/informers/externalversions"
)

const (
	defaultUpdateInterval = 60
	defaultLogLevel       = "debug"
	ptpNamespace	      = "ptp"
)

type cliParams struct {
	updateInterval	int
	logLevel	string
}

type Daemon struct {
        // name is the node name.
        nodeName      string
        namespace string

        ptpClient ptpclient.Interface
        // kubeClient allows interaction with Kubernetes, including the node we are running on.
        kubeClient *kubernetes.Clientset
}

func NewDaemon(
        nodeName string,
	namespace string,
        client ptpclient.Interface,
        kubeClient *kubernetes.Clientset,
) *Daemon {
        return &Daemon{
                nodeName:	nodeName,
		namespace:	namespace,
                ptpClient:	client,
                kubeClient:	kubeClient,
        }
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

func nodeLabelsGet(clientset *kubernetes.Clientset) (map[string]string, error) {
        nodeName := os.Getenv("PTP_NODE_NAME")
        if len(nodeName) > 0 {
                logging.Debugf("node name: %s", nodeName)
        } else {
                return nil, fmt.Errorf("Error getting node name, environment var PTP_NODE_NAME not set")
        }
        node, err := clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
        if errors.IsNotFound(err) {
                return nil, fmt.Errorf("Node %s not found", nodeName)
        } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
                return nil, fmt.Errorf("Error getting node %s: %v", nodeName, statusError.ErrStatus.Message)
        }
        if err != nil {
                return nil, err
        }

        return node.Labels, nil
}

func (dn *Daemon) updateNodePTPDevStatus(ptpDevs *ptpv1.NodePTPDev) {
	_, err := dn.ptpClient.PtpV1().NodePTPDevs(ptpNamespace).UpdateStatus(ptpDevs.Status)
	if err != nil {
		logging.Errorf("updateNodePTPDevStatus() failed: %v", err)
	}
}

func (dn *Daemon) nodePTPDevAdd(obj interface{}) {
	nodePTPDev := obj.(*ptpv1.NodePTPDev)
	logging.Debugf("nodePTPDevAdd(), nodePTPDev: %v", nodePTPDev)

	hostDevs, err := ptputils.DiscoverPTPDevices()
	if err != nil {
		logging.Debugf("discover PTP devices failed: %v", err)
		return
	}
	logging.Debugf("PTP capable NICs: %v", hostDevs)

	for _, dev := range hostDevs {
		nodePTPDev.Status.PTPDevices = append(nodePTPDev.Status.PTPDevices,
			ptpv1.PTPDevice{Name: dev, Profile: ""})
	}
	dn.updateNodePTPDevStatus(nodePTPDev)
}

func (dn *Daemon) nodePTPDevUpdate(oldStat, newStat interface{}) {
	oldNodePTPDev := oldStat.(*ptpv1.NodePTPDev)
	newNodePTPDev := newStat.(*ptpv1.NodePTPDev)

	if oldNodePTPDev.GetObjectMeta().GetGeneration() ==
		newNodePTPDev.GetObjectMeta().GetGeneration() { return }

	logging.Debugf("nodePTPDevUpdate(), oldNodePTPDev: %v", oldNodePTPDev)
	logging.Debugf("nodePTPDevUpdate(), newNodePTPDev: %v", newNodePTPDev)
}

func main() {
	cp := &cliParams{}
	flag.Parse()
	flagInit(cp)

	if cp.logLevel != ""{
		logging.SetLogLevel(cp.logLevel)
	}

	logging.Debugf("Resync period set to: %d [s]", cp.updateInterval)

	cfg, err := getConfig()
	if err != nil {
		logging.Errorf("get kubeconfig failed: %v", err)
		return
	}
	logging.Debugf("kubeconfig: %v", cfg)

        kubeClient, err := kubernetes.NewForConfig(cfg)
        if err != nil {
                logging.Debugf("cannot create new config for kubeClient: %v", err)
                return
        }

	ptpClient, err := ptpclient.NewForConfig(cfg)
	if err != nil {
		logging.Debugf("cannot create new config for ptpClient: %v", err)
		return
	}

	daemon := NewDaemon(os.Getenv("PTP_NODE_NAME"), ptpNamespace, ptpClient, kubeClient)
	logging.Debugf("daemon instance: %v", daemon)

	ptpInformerFactory := ptpinformer.NewFilteredSharedInformerFactory(
		ptpClient, time.Second*30, ptpNamespace,
                func(lo *metav1.ListOptions) {
                        lo.FieldSelector = "metadata.name=" + os.Getenv("PTP_NODE_NAME")
                },)
	ptpInformer := ptpInformerFactory.Ptp().V1().NodePTPDevs().Informer()
        ptpInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
                AddFunc:    daemon.nodePTPDevAdd,
                UpdateFunc: daemon.nodePTPDevUpdate,
        })

	time.Sleep(5 * time.Second)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go ptpInformer.Run(stopCh)

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

	for {
		select {
		case <-tickerPull.C:
			logging.Debugf("ticker pull")
			labels, err := nodeLabelsGet(kubeClient)
			if err != nil {
				logging.Debugf("get node labels failed: %v", err)
			} else {
				logging.Debugf("node labels: %v", labels)
			}
		case sig := <-sigCh:
			logging.Debugf("signal received, shutting down", sig)
			return
		}
	}
}
