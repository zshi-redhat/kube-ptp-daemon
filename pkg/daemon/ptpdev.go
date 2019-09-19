package daemon

import (
	"github.com/golang/glog"
	ptpnetwork "github.com/zshi-redhat/kube-ptp-daemon/pkg/network"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
)

func getDevStatusUpdate(nodePTPDev *ptpv1.NodePTPDev) (*ptpv1.NodePTPDev, error) {
	hostDevs, err := ptpnetwork.DiscoverPTPDevices()
	if err != nil {
		glog.Errorf("discover PTP devices failed: %v", err)
		return nodePTPDev, err
	}
	glog.Infof("PTP capable NICs: %v", hostDevs)
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
