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
	nodePTPDev := obj.(*ptpv1.NodePTPDev)
	glog.Infof("nodePTPDevAdd(), nodePTPDev: %v", nodePTPDev)

	update, err := getDevStatusUpdate(nodePTPDev)
	if err != nil {
		glog.Errorf("getDevStatusUpdate failed: %v", err)
		return
	}
	dn.updateNodePTPDevStatus(update)
}

func (dn *Daemon) nodePTPDevUpdateHandler(oldStat, newStat interface{}) {
	oldNodePTPDev := oldStat.(*ptpv1.NodePTPDev)
	newNodePTPDev := newStat.(*ptpv1.NodePTPDev)

	if oldNodePTPDev.GetObjectMeta().GetGeneration() ==
		newNodePTPDev.GetObjectMeta().GetGeneration() { return }

	glog.Infof("nodePTPDevUpdate(), oldNodePTPDev: %v", oldNodePTPDev)
	glog.Infof("nodePTPDevUpdate(), newNodePTPDev: %v", newNodePTPDev)

	update, err := getDevStatusUpdate(newNodePTPDev)
	if err != nil {
		glog.Errorf("getDevStatusUpdate failed: %v", err)
	}
	dn.updateNodePTPDevStatus(update)
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
	_, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).Get(dn.nodeName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		createdPTPDev, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).Create(ptpDev)
		if err != nil {
			glog.Errorf("createNodePTPDevResource() failed: %v", err)
			return
		}
		glog.Infof("createNodePTPDevResource(), resource successfull created: %v", createdPTPDev)
        } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
                glog.Errorf("Error getting nodePTPDev %s: %v", dn.nodeName, statusError.ErrStatus.Message)
		return
        } else {
		glog.Infof("createNodePTPDevResource(), resource already exist, skipping")
	}
}

func (dn *Daemon) updateNodePTPDevStatus(ptpDev *ptpv1.NodePTPDev) {
	updatedPTPDev, err := dn.ptpClient.PtpV1().NodePTPDevs(PtpNamespace).UpdateStatus(ptpDev)
	if err != nil {
		glog.Errorf("updateNodePTPDevStatus() failed: %v", err)
	}
	glog.Infof("updateNodePTPDevStatus(), status successfull updated to: %v", updatedPTPDev.Status)
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

func (dn *Daemon) updateLinuxPTPInstance(confList *ptpv1.NodePTPCfgList) {
	nodeLabels, err := dn.getNodeLabels(dn.kubeClient)
	if err != nil {
		glog.Infof("get node labels failed: %v", err)
		return
	}
	glog.Infof("updateLinuxPTPInstance() node labels: %+v", nodeLabels)

	nodeCfgUpdate, err := getNodePTPCfgUpdate(confList, dn.nodeName, nodeLabels)
	if err != nil {
		glog.Errorf("get nodePTPCfgToUpdate failed: %v", err)
		return
	}
	glog.Infof("getNodePTPCfgUpdate() nodeCfgUpdate :+v", nodeCfgUpdate)
	dn.updateNodePTPCfgStatus(nodeCfgUpdate.current, nodeCfgUpdate.update)

	dn.ptpUpdate.nodeProfile = nodeCfgUpdate.nodeProfile
	dn.ptpUpdate.updateCh <- true
}

func (dn *Daemon) nodePTPCfgAddHandler(obj interface{}) {
	nodePTPCfg := obj.(*ptpv1.NodePTPCfg)
	glog.Infof("nodePTPCfgAdd(), nodePTPCfg: %+v", nodePTPCfg)

	confList, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("failed to list NodePTPCfgs: %v", err)
		return
	}
	dn.updateLinuxPTPInstance(confList)
}

func (dn *Daemon) nodePTPCfgUpdateHandler(oldStat, newStat interface{}) {
	oldNodePTPCfg := oldStat.(*ptpv1.NodePTPCfg)
	newNodePTPCfg := newStat.(*ptpv1.NodePTPCfg)

	if oldNodePTPCfg.GetObjectMeta().GetGeneration() ==
		newNodePTPCfg.GetObjectMeta().GetGeneration() { return }

	glog.Infof("nodePTPCfgUpdate(), oldNodePTPCfg: %v", oldNodePTPCfg)
	glog.Infof("nodePTPCfgUpdate(), newNodePTPCfg: %v", newNodePTPCfg)

	confList, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("failed to list NodePTPCfgs: %v", err)
		return
	}
	dn.updateLinuxPTPInstance(confList)
}

func (dn *Daemon) updateNodePTPCfgStatus(current, update ptpv1.NodePTPCfg) {
	if current.Name != "" && current.Name != update.Name {
		updatedCfg, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).UpdateStatus(&current)
		if err != nil {
			glog.Errorf("updateNodePTPCfgStatus() current failed: %v", err)
			return
		}
		glog.Infof("updateNodePTPCfgStatus() Current successfully: %+v", updatedCfg)
	}

	if update.Name != "" {
		updatedCfg, err := dn.ptpClient.PtpV1().NodePTPCfgs(PtpNamespace).UpdateStatus(&update)
		if err != nil {
			glog.Errorf("updateNodePTPCfgStatus() update failed: %v", err)
			return
		}
		glog.Infof("updateNodePTPCfgStatus() Update successfully: %+v", updatedCfg)
	}
	glog.Infof("updateNodePTPCfgStatus() Update nothing")
}
