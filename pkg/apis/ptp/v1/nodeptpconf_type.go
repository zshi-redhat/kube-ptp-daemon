package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodePTPConfSpec defines the desired state of NodePTPConf
type NodePTPConfSpec struct {
	Profile		[]NodePTPProfile	`json:"profile"`
	Recommend	[]NodePTPRecommend	`json:"recommend"`
}

type NodePTPProfile struct {
	Name		string		`json:"name"`
	Interfaces	[]string	`json:"interfaces"`
	Ptp4lOpts	string		`json:"ptp4lOpts,omitempty"`
	Phc2sysOpts	string		`json:"phc2sysOpts,omitempty"`
	Ptp4lConf	*string		`json:"ptp4lConf,omitempty"`
}

type NodePTPRecommend struct {
	Profile		string			`json:"profile"`
	Priority	int64			`json:"priority"`
	Match		[]NodePTPMatchRule	`json:"match,omitempty"`
}

type NodePTPMatchRule struct {
	NodeLabel	string	`json:"nodeLabel,omitempty"`
	NodeName	string	`json:"nodeName,omitempty"`
}

// NodePTPConfStatus defines the observed state of NodePTPConf
type NodePTPConfStatus struct {
	MatchList	[]NodeMatchList	`json:"matchList,omitempty"`
}

type NodeMatchList struct {
	NodeName	string	`json:"nodeName"`
	Profile		string	`json:"profile"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodePTPConf is the Schema for the nodeptpconfs API
// +kubebuilder:subresource:status
type NodePTPConf struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        Spec   NodePTPConfSpec   `json:"spec,omitempty"`
        Status NodePTPConfStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodePTPConfList contains a list of NodePTPConf
type NodePTPConfList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata,omitempty"`
        Items           []NodePTPConf `json:"items"`
}
