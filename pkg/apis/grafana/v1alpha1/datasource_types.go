package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataSource is a specification for a DataSource resource
type DataSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSourceSpec   `json:"spec"`
	Status DataSourceStatus `json:"status"`
}

// DataSourceSpec is the spec for a DataSource resource
type DataSourceSpec struct {
	JSON string `json:"json"`
}

// DataSourceStatus is the status for a DataSource resource
type DataSourceStatus struct {
	GrafanaID string `json:"grafanaID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataSourceList is a list of DataSource resources
type DataSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DataSource `json:"items"`
}
