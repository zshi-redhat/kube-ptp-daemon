package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodePTPCfgSpec defines the desired state of NodePTPCfg
type NodePTPCfgSpec struct {
	Profile		[]NodePTPProfile	`json:"profile"`
	Recommend	[]NodePTPRecommend	`json:"recommend"`
}

type NodePTPProfile struct {
	Name		*string		`json:"name"`
	Interface	*string		`json:"interface"`
	Ptp4lOpts	*string		`json:"ptp4lOpts,omitempty"`
	Phc2sysOpts	*string		`json:"phc2sysOpts,omitempty"`
	Ptp4lConf	*string		`json:"ptp4lConf,omitempty"`
}

type NodePTPRecommend struct {
	Profile		*string			`json:"profile"`
	Priority	*int64			`json:"priority"`
	Match		[]NodePTPMatchRule	`json:"match,omitempty"`
}

type NodePTPMatchRule struct {
	NodeLabel	*string	`json:"nodeLabel,omitempty"`
	NodeName	*string	`json:"nodeName,omitempty"`
}

// NodePTPCfgStatus defines the observed state of NodePTPCfg
type NodePTPCfgStatus struct {
	MatchList	[]NodeMatchList	`json:"matchList,omitempty"`
}

type NodeMatchList struct {
	NodeName	*string	`json:"nodeName"`
	Profile		*string	`json:"profile"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodePTPCfg is the Schema for the nodeptpconfs API
// +kubebuilder:subresource:status
type NodePTPCfg struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        Spec   NodePTPCfgSpec   `json:"spec,omitempty"`
        Status NodePTPCfgStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodePTPCfgList contains a list of NodePTPCfg
type NodePTPCfgList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata,omitempty"`
        Items           []NodePTPCfg `json:"items"`
}
