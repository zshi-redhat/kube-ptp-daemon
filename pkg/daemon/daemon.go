package daemon

import(
	"fmt"
	"time"

	"github.com/golang/glog"
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
        nodeName	string
        namespace	string

        ptpClient	ptpclient.Interface
        // kubeClient allows interaction with Kubernetes, including the node we are running on.
        kubeClient	*kubernetes.Clientset

	ptpUpdate	*linuxPTPUpdate
	// channel ensure daemon.Run() exit when main function exits.
	// stopCh is created by main function and passed to Daemon via daemon.New()
	stopCh		<-chan struct{}
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
		ptpUpdate:	&linuxPTPUpdate{updateCh: make(chan bool)},
		stopCh:		stopCh,
        }
}

func (dn *Daemon) Run() error {
	glog.V(2).Infof("in Run")

	go NewLinuxPTP(
		dn.nodeName,
		dn.namespace,
		dn.ptpUpdate,
		dn.stopCh,
	).Run()

	ptpInformerFactory := ptpinformer.NewFilteredSharedInformerFactory(
		dn.ptpClient, time.Second*30, PtpNamespace,
                func(lo *metav1.ListOptions) {
//                        lo.FieldSelector = "metadata.name=" + os.Getenv("PTP_NODE_NAME")
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
	glog.V(2).Infof("in nodePTPDevAddHandler")

	nodePTPDev := obj.(*ptpv1.NodePTPDev)
	glog.Infof("new nodePTPDev CR is created: %v", nodePTPDev)

	update, err := getDevStatusUpdate(nodePTPDev)
	if err != nil {
		glog.Errorf("get PTP device information failed: %v", err)
		return
	}
	glog.Infof("get PTP device information succeeded")
	dn.updateNodePTPDevStatus(update)
}

func (dn *Daemon) nodePTPDevUpdateHandler(oldStat, newStat interface{}) {
	glog.V(2).Infof("in nodePTPDevUpdateHandler")

	oldNodePTPDev := oldStat.(*ptpv1.NodePTPDev)
	newNodePTPDev := newStat.(*ptpv1.NodePTPDev)

	if oldNodePTPDev.GetObjectMeta().GetGeneration() ==
		newNodePTPDev.GetObjectMeta().GetGeneration() { return }

	glog.Infof("nodePTPDev CR is updated, oldNodePTPDev: %v", oldNodePTPDev)
	glog.Infof("nodePTPDev CR is updated, newNodePTPDev: %v", newNodePTPDev)

	update, err := getDevStatusUpdate(newNodePTPDev)
	if err != nil {
		glog.Errorf("get PTP device information failed: %v", err)
	}
	dn.updateNodePTPDevStatus(update)
}

func (dn *Daemon) createNodePTPDevResource() {
	glog.V(2).Infof("in createNodePTPDevResource")

	ptpDev := &ptpv1.NodePTPDev{
			ObjectMeta: metav1.ObjectMeta{
				Name: dn.nodeName,
				Namespace: PtpNamespace,
			},
			Spec: ptpv1.NodePTPDevSpec{
				PTPDevices: []ptpv1.PTPDevice{},
			},
		}
	_, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).Get(dn.nodeName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		createdPTPDev, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).Create(ptpDev)
		if err != nil {
			glog.Errorf("create NodePTPDev CR failed: %v", err)
			return
		}
		glog.Infof("create NodePTPDev CR succeeded: %v", createdPTPDev)
        } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
                glog.Errorf("Error getting nodePTPDev %s: %v", dn.nodeName, statusError.ErrStatus.Message)
		return
        } else {
		glog.Infof("creation of NodePTPDev CR skipped, resource already exist")
	}
}

func (dn *Daemon) updateNodePTPDevStatus(ptpDev *ptpv1.NodePTPDev) {
	glog.V(2).Infof("in updateNodePTPDevStatus")

	updatedPTPDev, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).UpdateStatus(ptpDev)
	if err != nil {
		glog.Errorf("update NodePTPDev Status failed for node %v: %v", dn.nodeName, err)
	}
	glog.Infof("successfull update NodePTPDev Status to: %v", updatedPTPDev.Status)
}

func (dn *Daemon) getNodeLabels(clientset *kubernetes.Clientset) (map[string]string, error) {
	glog.V(2).Infof("in getNodeLabels")

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

func (dn *Daemon) updateLinuxPTPInstance(confList *ptpv1.NodePTPCfgList) {
	glog.V(2).Infof("in updateLinuxPTPInstance")

	nodeLabels, err := dn.getNodeLabels(dn.kubeClient)
	if err != nil {
		glog.Infof("get node labels failed: %v", err)
		return
	}
	glog.Infof("node labels: %+v", nodeLabels)

	nodeCfgUpdate, err := getNodePTPCfgUpdate(confList, dn.nodeName, nodeLabels)
	if err != nil {
		glog.Errorf("get updated NodePTPCfg CR failed: %v", err)
		return
	}
	glog.Infof("get updated NodePTPCfg CR succeeded: %+v", nodeCfgUpdate)
	dn.updateNodePTPCfgStatus(nodeCfgUpdate.current, nodeCfgUpdate.update)

	dn.ptpUpdate.nodeProfile = nodeCfgUpdate.nodeProfile
	dn.ptpUpdate.updateCh <- true
}

func (dn *Daemon) nodePTPCfgAddHandler(obj interface{}) {
	glog.V(2).Infof("in nodePTPCfgAddHandler")

	nodePTPCfg := obj.(*ptpv1.NodePTPCfg)
	glog.Infof("new nodePTPCfg CR is created: %+v", nodePTPCfg)

	confList, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("failed to list NodePTPCfg CRs: %v", err)
		return
	}
	dn.updateLinuxPTPInstance(confList)
}

func (dn *Daemon) nodePTPCfgUpdateHandler(oldStat, newStat interface{}) {
	glog.V(2).Infof("in nodePTPCfgUpdateHandler")

	oldNodePTPCfg := oldStat.(*ptpv1.NodePTPCfg)
	newNodePTPCfg := newStat.(*ptpv1.NodePTPCfg)

	if oldNodePTPCfg.GetObjectMeta().GetGeneration() ==
		newNodePTPCfg.GetObjectMeta().GetGeneration() { return }

	glog.Infof("nodePTPCfg CR is updated, oldNodePTPCfg: %v", oldNodePTPCfg)
	glog.Infof("nodePTPCfg CR is updated, newNodePTPCfg: %v", newNodePTPCfg)

	confList, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("failed to list NodePTPCfgs: %v", err)
		return
	}
	dn.updateLinuxPTPInstance(confList)
}

func (dn *Daemon) updateNodePTPCfgStatus(current, update ptpv1.NodePTPCfg) {
	glog.V(2).Infof("in updateNodePTPCfgStatus")

	if current.Name != "" && current.Name != update.Name {
		updatedCfg, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).UpdateStatus(&current)
		if err != nil {
			glog.Errorf("update 'current' NodePTPCfg Status failed: %v", err)
			return
		}
		glog.Infof("update 'current' NodePTPCfg Status succeeded: %+v", updatedCfg)
	}

	if update.Name != "" {
		updatedCfg, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).UpdateStatus(&update)
		if err != nil {
			glog.Errorf("update 'update' NodePTPCfg Status failed: %v", err)
			return
		}
		glog.Infof("update 'update' NodePTPCfg Status succeeded: %+v", updatedCfg)
	}
}
