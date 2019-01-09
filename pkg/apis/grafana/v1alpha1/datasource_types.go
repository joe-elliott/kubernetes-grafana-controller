package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDataSource is a specification for a GrafanaDataSource resource
type GrafanaDataSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDataSourceSpec   `json:"spec"`
	Status GrafanaDataSourceStatus `json:"status"`
}

// GrafanaDataSourceSpec is the spec for a GrafanaDataSource resource
type GrafanaDataSourceSpec struct {
	DataSourceJSON string `json:"dataSourceJson"`
}

// GrafanaDataSourceStatus is the status for a GrafanaDataSource resource
type GrafanaDataSourceStatus struct {
	GrafanaID string `json:"grafanaID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDataSourceList is a list of GrafanaDataSource resources
type GrafanaDataSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []GrafanaDataSource `json:"items"`
}
