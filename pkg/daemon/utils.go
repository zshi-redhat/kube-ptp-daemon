package daemon

import (
	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptputils "github.com/zshi-redhat/kube-ptp-daemon/pkg/utils"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
)

func getDevStatusUpdate(nodePTPDev *ptpv1.NodePTPDev) (*ptpv1.NodePTPDev, error) {
	hostDevs, err := ptputils.DiscoverPTPDevices()
	if err != nil {
		logging.Errorf("discover PTP devices failed: %v", err)
		return nodePTPDev, err
	}
	logging.Debugf("PTP capable NICs: %v", hostDevs)
	for _, hostDev := range hostDevs {
		contained := false
		for _, crDev := range nodePTPDev.Status.PTPDevices {
			if hostDev == crDev.Name {
				contained = true
				break
			}
		}
		if !contained {
			nodePTPDev.Status.PTPDevices = append(nodePTPDev.Status.PTPDevices,
				ptpv1.PTPDevice{Name: hostDev, Profile: ""})
		}
	}
	return nodePTPDev, nil
}
