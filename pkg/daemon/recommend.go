package daemon

import (
	"sort"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
)

func getCfgUpdate(
	ptpCfgList *ptpv1.NodePTPCfgList,
	nodeName string,
	nodeLabels map[string]string,
) (
	ptpv1.NodePTPCfg,
	ptpv1.NodePTPCfg,
	error,
) {
	var (
		cfgCurrent ptpv1.NodePTPCfg
		cfgToUpdate ptpv1.NodePTPCfg
	)

	profileName, _ := getRecommendProfile(ptpCfgList, nodeName, nodeLabels)
	logging.Debugf("getNodePTPCfgToUpdate(), profile name: %+v", profileName)
	for _, cfg := range ptpCfgList.Items {
		if cfg.Status.MatchList != nil {
			for idx, m := range cfg.Status.MatchList {
				if *m.NodeName == nodeName {
					cfg.Status.MatchList = append(
						cfg.Status.MatchList[:idx],
						cfg.Status.MatchList[idx+1:]...)
					logging.Debugf("getNodePTPCfgToUpdate(), append current cfg: %+v", cfg)
					cfgCurrent = cfg
				}
			}
		}
		if cfg.Spec.Profile != nil {
			for _, p := range cfg.Spec.Profile {
				if profileName == *p.Name {
					cfg.Status.MatchList = append(
						cfg.Status.MatchList,
						ptpv1.NodeMatchList{
							NodeName: &nodeName,
							Profile: &profileName})
					logging.Debugf("getNodePTPCfgToUpdate(), append toUpdate cfg: %+v", cfg)
					cfgToUpdate = cfg
				}
			}
		}
	}
	logging.Debugf("getNodePTPCfgToUpdate(), cfgCurrent: %+v", cfgCurrent)
	logging.Debugf("getNodePTPCfgToUpdate(), cfgToUpdate: %+v", cfgToUpdate)
	return cfgCurrent, cfgToUpdate, nil
}

func getRecommendProfile(
	ptpCfgList *ptpv1.NodePTPCfgList,
	nodeName string,
	nodeLabels map[string]string,
) ( string, error ) {
	var (
		labelMatches	[]string
		allRecommend	[]ptpv1.NodePTPRecommend
	)

	// append recommend section from each custom resource into one list
	for _, cfg := range ptpCfgList.Items {
		logging.Debugf("getRecommendProfileName(), nodePTPCfg: %+v", cfg)
		if cfg.Spec.Recommend != nil {
			allRecommend = append(allRecommend, cfg.Spec.Recommend...)
		}
	}

	// allRecommend sorted by priority
	// priority 0 will become the first item in allRecommend
	sort.Slice(allRecommend, func(i, j int) bool {
		if allRecommend[i].Priority != nil && allRecommend[j].Priority != nil {
			return *allRecommend[i].Priority < *allRecommend[j].Priority
		}
		return allRecommend[i].Priority != nil
	})

	// loop allRecommend from high priority(0) to low(*)
	for _, r := range allRecommend {

		// ignore if profile not define in recommend
		if r.Profile != nil {

			// ignore if match section is empty
			if len(r.Match) == 0 {
				continue
			}

			// loop over Match list
			for _, m := range r.Match {

				// nodeName has higher priority than nodeLabel
				// return immediately if nodeName matches
				if *m.NodeName == nodeName {
					return *r.Profile, nil
				}

				// don't return immediately when label matches
				// chance is next Match item may hit NodeName
				for k, _ := range nodeLabels {
					return *r.Profile, nil
					if *m.NodeLabel == k {
						labelMatches = append(labelMatches, *r.Profile)
					}
				}
			}
			if len(labelMatches) > 0 {
				break
			}
		}
	}
	return "", nil
}
