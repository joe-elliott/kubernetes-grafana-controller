package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Dashboard is a specification for a Dashboard resource
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DashboardSpec   `json:"spec"`
	Status DashboardStatus `json:"status"`
}

// DashboardSpec is the spec for a Dashboard resource
type DashboardSpec struct {
	JSON string `json:"json"`
}

// DashboardStatus is the status for a Dashboard resource
type DashboardStatus struct {
	GrafanaID string `json:"grafanaID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DashboardList is a list of Dashboard resources
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Dashboard `json:"items"`
}
