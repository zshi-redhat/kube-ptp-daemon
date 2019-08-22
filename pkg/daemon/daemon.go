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
	ptpDevInformer := ptpInformerFactory.Ptp().V1().NodePTPDevs().Informer()
        ptpDevInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
                AddFunc:    dn.nodePTPDevAddHandler,
                UpdateFunc: dn.nodePTPDevUpdateHandler,
        })

	ptpCfgInformer := ptpInformerFactory.Ptp().V1().NodePTPCfgs().Informer()
        ptpCfgInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
                AddFunc:    dn.nodePTPCfgAddHandler,
                UpdateFunc: dn.nodePTPCfgUpdateHandler,
        })

	time.Sleep(2 * time.Second)
	go ptpDevInformer.Run(dn.stopCh)
	go ptpCfgInformer.Run(dn.stopCh)

	time.Sleep(2 * time.Second)
	// create per-node resource ptpv1.NodePTPDev
	dn.createNodePTPDevResource()
	return nil
}

func (dn *Daemon) nodePTPDevAddHandler(obj interface{}) {
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

func (dn *Daemon) nodePTPDevUpdateHandler(oldStat, newStat interface{}) {
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

func (dn *Daemon) createNodePTPDevResource() {
	ptpDev := &ptpv1.NodePTPDev{
			ObjectMeta: metav1.ObjectMeta{
				Name: dn.nodeName,
				Namespace: PtpNamespace,
			},
			Spec: ptpv1.NodePTPDevSpec{
				PTPDevices: []ptpv1.PTPDevice{},
			},
		}
	createdPTPDev, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).Create(ptpDev)
	if err != nil {
		logging.Errorf("createNodePTPDevResource() failed: %v", err)
	}
	logging.Debugf("createNodePTPDevResource(), resource successfull created: %v", createdPTPDev)
}

func (dn *Daemon) updateNodePTPDevStatus(ptpDev *ptpv1.NodePTPDev) {
	updatedPTPDev, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).UpdateStatus(ptpDev)
	if err != nil {
		logging.Errorf("updateNodePTPDevStatus() failed: %v", err)
	}
	logging.Debugf("updateNodePTPDevStatus(), status successfull updated to: %v", updatedPTPDev.Status)
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

func (dn *Daemon) nodePTPCfgAddHandler(obj interface{}) {
	nodePTPCfg := obj.(*ptpv1.NodePTPCfg)
	logging.Debugf("nodePTPCfgAdd(), nodePTPCfg: %v", nodePTPCfg)

	confList, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).List(metav1.ListOptions{})
	if err != nil {
		logging.Errorf("failed to list NodePTPCfgs: %v", err)
		return
	}
	for _, conf := range confList.Items {
		logging.Debugf("nodePTPCfgAddHandler(), nodePTPCfg: %+v", conf)
	}
}

func (dn *Daemon) nodePTPCfgUpdateHandler(oldStat, newStat interface{}) {
	oldNodePTPCfg := oldStat.(*ptpv1.NodePTPCfg)
	newNodePTPCfg := newStat.(*ptpv1.NodePTPCfg)

	if oldNodePTPCfg.GetObjectMeta().GetGeneration() ==
		newNodePTPCfg.GetObjectMeta().GetGeneration() { return }

	logging.Debugf("nodePTPCfgUpdate(), oldNodePTPCfg: %v", oldNodePTPCfg)
	logging.Debugf("nodePTPCfgUpdate(), newNodePTPCfg: %v", newNodePTPCfg)

	nodeLabels, err := dn.getNodeLabels(dn.kubeClient)
	if err != nil {
		logging.Debugf("get node labels failed: %v", err)
	} else {
		logging.Debugf("node labels: %v", nodeLabels)
	}
}

func (dn *Daemon) updateNodePTPCfgStatus(ptpCfg *ptpv1.NodePTPCfg) {
	updatedPTPCfg, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).UpdateStatus(ptpCfg)
	if err != nil {
		logging.Errorf("updateNodePTPCfgStatus() failed: %v", err)
	}
	logging.Debugf("updateNodePTPCfgStatus(), config status successfull updated to: %v",
		updatedPTPCfg.Status)
}

