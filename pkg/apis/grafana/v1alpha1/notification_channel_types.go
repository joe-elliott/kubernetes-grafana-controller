package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertNotification is a specification for a AlertNotification resource
type AlertNotification struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertNotificationSpec   `json:"spec"`
	Status AlertNotificationStatus `json:"status"`
}

// AlertNotificationSpec is the spec for a AlertNotification resource
type AlertNotificationSpec struct {
	JSON string `json:"json"`
}

// AlertNotificationStatus is the status for a AlertNotification resource
type AlertNotificationStatus struct {
	GrafanaID string `json:"grafanaID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertNotificationList is a list of AlertNotification resources
type AlertNotificationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AlertNotification `json:"items"`
}
