package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NotificationChannel is a specification for a NotificationChannel resource
type NotificationChannel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotificationChannelSpec   `json:"spec"`
	Status NotificationChannelStatus `json:"status"`
}

// NotificationChannelSpec is the spec for a NotificationChannel resource
type NotificationChannelSpec struct {
	JSON string `json:"json"`
}

// NotificationChannelStatus is the status for a NotificationChannel resource
type NotificationChannelStatus struct {
	GrafanaID string `json:"grafanaID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NotificationChannelList is a list of NotificationChannel resources
type NotificationChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NotificationChannel `json:"items"`
}
