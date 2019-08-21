package daemon

import(
	"fmt"
	"os"
	"time"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptputils "github.com/zshi-redhat/kube-ptp-daemon/pkg/utils"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
	ptpclient "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/clientset/versioned"
	ptpinformer "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/informers/externalversions"
        "k8s.io/apimachinery/pkg/api/errors"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	PtpNamespace = "ptp"
)

type Daemon struct {
        // name is the node name.
        nodeName      string
        namespace string

        ptpClient ptpclient.Interface
        // kubeClient allows interaction with Kubernetes, including the node we are running on.
        kubeClient *kubernetes.Clientset

	// channel ensure daemon.Run() exit when main function exits.
	// stopCh is created by main function and passed to Daemon via daemon.New()
	stopCh <-chan struct{}
}

func New(
        nodeName string,
	namespace string,
        client ptpclient.Interface,
        kubeClient *kubernetes.Clientset,
	stopCh <-chan struct{},
) *Daemon {
        return &Daemon{
                nodeName:	nodeName,
		namespace:	namespace,
                ptpClient:	client,
                kubeClient:	kubeClient,
		stopCh:		stopCh,
        }
}

func (dn *Daemon) Run() error {
	ptpInformerFactory := ptpinformer.NewFilteredSharedInformerFactory(
		dn.ptpClient, time.Second*30, PtpNamespace,
                func(lo *metav1.ListOptions) {
                        lo.FieldSelector = "metadata.name=" + os.Getenv("PTP_NODE_NAME")
                },)
	ptpInformer := ptpInformerFactory.Ptp().V1().NodePTPDevs().Informer()
        ptpInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
                AddFunc:    dn.nodePTPDevAdd,
                UpdateFunc: dn.nodePTPDevUpdate,
        })

	time.Sleep(5 * time.Second)
	go ptpInformer.Run(dn.stopCh)
	return nil
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

	nodeLabels, err := dn.getNodeLabels(dn.kubeClient)
	if err != nil {
		logging.Debugf("get node labels failed: %v", err)
	} else {
		logging.Debugf("node labels: %v", nodeLabels)
	}
}

func (dn *Daemon) updateNodePTPDevStatus(ptpDevs *ptpv1.NodePTPDev) {
	_, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).UpdateStatus(ptpDevs)
	if err != nil {
		logging.Errorf("updateNodePTPDevStatus() failed: %v", err)
	}
}

func (dn *Daemon) getNodeLabels(clientset *kubernetes.Clientset) (map[string]string, error) {
        node, err := clientset.CoreV1().Nodes().Get(dn.nodeName, metav1.GetOptions{})
        if errors.IsNotFound(err) {
                return nil, fmt.Errorf("Node %s not found", dn.nodeName)
        } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
                return nil, fmt.Errorf("Error getting node %s: %v", dn.nodeName, statusError.ErrStatus.Message)
        }
        if err != nil {
                return nil, err
        }
        return node.Labels, nil
}
