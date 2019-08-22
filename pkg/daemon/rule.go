package daemon

import (
	"sort"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
)

func getRecommendProfileName(
	ptpCfgList *ptpv1.NodePTPCfgList,
	nodeName string,
	nodeLabels map[string]string,
) ( string, error ) {
	var (
		allRecommend	[]ptpv1.NodePTPRecommend
	)

	for _, cfg := range ptpCfgList.Items {
		logging.Debugf("nodePTPCfgAddHandler(), nodePTPCfg: %+v", cfg)
		if cfg.Spec.Recommend != nil {
			allRecommend = append(allRecommend, cfg.Spec.Recommend...)
		}
	}

	sort.Slice(allRecommend, func(i, j int) bool {
		if allRecommend[i].Priority != nil && allRecommend[j].Priority != nil {
			return *allRecommend[i].Priority < *allRecommend[j].Priority
		}
		return allRecommend[i].Priority != nil
	})

	for _, r := range allRecommend {
		if r.Profile != nil {
			if len(r.Match) == 0 {
				continue
			}
			for _, m := range r.Match {
				if *m.NodeName == nodeName {
					return *r.Profile, nil
				}
				for k, _ := range nodeLabels {
					if *m.NodeLabel == k {
						return *r.Profile, nil
					}
				}
			}
		}
	}
	return "", nil
}
