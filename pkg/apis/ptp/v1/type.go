package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodePTPDevSpec defines the desired state of NodePTPDev
type NodePTPDevSpec struct {
}

type PTPDevice struct {
	Name		string	`json:"name"`
	Profile		string	`json:"profile"`
}

// NodePTPDevStatus defines the observed state of NodePTPDev
type NodePTPDevStatus struct {
	PTPDevices	[]PTPDevice	`json:"ptpDevices"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodePTPDev is the Schema for the nodeptpdevs API
// +kubebuilder:subresource:status
type NodePTPDev struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        Spec   NodePTPDevSpec   `json:"spec,omitempty"`
        Status NodePTPDevStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodePTPDevList contains a list of NodePTPDev
type NodePTPDevList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata,omitempty"`
        Items           []NodePTPDev `json:"items"`
}
