package daemon

import (
	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
)

type linuxPTPUpdate struct {
	updateCh	chan bool
	nodeProfile	*ptpv1.NodePTPProfile
}

type linuxPTP struct {
	// node name where daemon is running
	nodeName	string
	namespace	string

	ptpUpdate	*linuxPTPUpdate
	// channel ensure LinuxPTP.Run() exit when main function exits.
	// stopCh is created by main function and passed by Daemon via NewLinuxPTP()
	stopCh <-chan struct{}
}

func NewLinuxPTP(
	nodeName	string,
	namespace	string,
	ptpUpdate	*linuxPTPUpdate,
	stopCh		<-chan struct{},
) *linuxPTP {
	return &linuxPTP{
		nodeName:	nodeName,
		namespace:	namespace,
		ptpUpdate:	ptpUpdate,
		stopCh:		stopCh,
	}
}

func (lp *linuxPTP) Run() {
	for {
		select {
		case <-lp.ptpUpdate.updateCh:
			err := lp.applyNodePTPProfile(lp.ptpUpdate.nodeProfile)
			if err != nil {
				logging.Errorf("linuxPTP apply node profile failed: %v", err)
			}
		case <-lp.stopCh:
			logging.Debugf("linuxPTP stop signal received, existing..")
			return
		}
	}
	return
}

func (lp *linuxPTP)applyNodePTPProfile(nodeProfile *ptpv1.NodePTPProfile) error {
	logging.Debugf("applyNodePTPProfile() NodePTPProfile: %+v", nodeProfile)
	return nil
}

