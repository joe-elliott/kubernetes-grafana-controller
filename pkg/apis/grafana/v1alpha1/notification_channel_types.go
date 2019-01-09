package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaNotificationChannel is a specification for a GrafanaNotificationChannel resource
type GrafanaNotificationChannel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaNotificationChannelSpec   `json:"spec"`
	Status GrafanaNotificationChannelStatus `json:"status"`
}

// GrafanaNotificationChannelSpec is the spec for a GrafanaNotificationChannel resource
type GrafanaNotificationChannelSpec struct {
	NotificationChannelJSON string `json:"notificationChannelJson"`
}

// GrafanaNotificationChannelStatus is the status for a GrafanaNotificationChannel resource
type GrafanaNotificationChannelStatus struct {
	GrafanaID string `json:"grafanaID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaNotificationChannelList is a list of GrafanaNotificationChannel resources
type GrafanaNotificationChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []GrafanaNotificationChannel `json:"items"`
}
