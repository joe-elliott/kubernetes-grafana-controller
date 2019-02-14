package controllers

import (
	"fmt"
	"kubernetes-grafana-controller/pkg/prometheus"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	clientset "kubernetes-grafana-controller/pkg/client/clientset/versioned"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions/grafana/v1alpha1"
	listers "kubernetes-grafana-controller/pkg/client/listers/grafana/v1alpha1"
	"kubernetes-grafana-controller/pkg/grafana"
)

// AlertNotificationSyncer is the controller implementation for GrafanaAlertNotification resources
type AlertNotificationSyncer struct {
	grafanaAlertNotificationLister listers.AlertNotificationLister
	grafanaClient                  grafana.Interface
	grafanaclientset               clientset.Interface
}

// NewAlertNotificationController returns a new grafana alert notification controller
func NewAlertNotificationController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaAlertNotificationInformer informers.AlertNotificationInformer) *Controller {

	syncer := &AlertNotificationSyncer{
		grafanaAlertNotificationLister: grafanaAlertNotificationInformer.Lister(),
		grafanaClient:                  grafanaClient,
		grafanaclientset:               grafanaclientset,
	}

	controller := NewController(grafanaAlertNotificationInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

func (s *AlertNotificationSyncer) getType() string {
	return prometheus.TypeAlertNotification
}

func (s *AlertNotificationSyncer) getRuntimeObjectByName(name string, namespace string) (runtime.Object, error) {
	return s.grafanaAlertNotificationLister.AlertNotifications(namespace).Get(name)
}

func (s *AlertNotificationSyncer) deleteObjectById(id string) error {
	return s.grafanaClient.DeleteAlertNotification(id)
}

func (s *AlertNotificationSyncer) updateObject(object runtime.Object) error {

	grafanaAlertNotification, ok := object.(*v1alpha1.AlertNotification)
	if !ok {
		return fmt.Errorf("expected alert notification in but got %#v", object)
	}

	id, err := s.grafanaClient.PostAlertNotification(grafanaAlertNotification.Spec.JSON, grafanaAlertNotification.Status.GrafanaID)

	if err != nil {
		return err
	}

	grafanaAlertNotificationCopy := grafanaAlertNotification.DeepCopy()
	grafanaAlertNotificationCopy.Status.GrafanaID = id
	_, err = s.grafanaclientset.GrafanaV1alpha1().AlertNotifications(grafanaAlertNotification.Namespace).UpdateStatus(grafanaAlertNotificationCopy)

	if err != nil {
		return err
	}
	return nil
}

func (s *AlertNotificationSyncer) getAllKubernetesObjectIDs() ([]string, error) {
	alertNotifications, err := s.grafanaAlertNotificationLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)

	for _, notification := range alertNotifications {
		ids = append(ids, notification.Status.GrafanaID)
	}

	return ids, nil
}

func (s *AlertNotificationSyncer) getAllGrafanaObjectIDs() ([]string, error) {
	return s.grafanaClient.GetAllAlertNotificationIds()
}

func (s *AlertNotificationSyncer) createWorkQueueItem(obj interface{}) *WorkQueueItem {
	var key string
	var err error
	var grafanaAlertNotification *v1alpha1.AlertNotification
	var ok bool

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return nil
	}

	if grafanaAlertNotification, ok = obj.(*v1alpha1.AlertNotification); !ok {
		utilruntime.HandleError(fmt.Errorf("expected AlertNotification in workqueue but got %#v", obj))
		return nil
	}

	item := NewWorkQueueItem(key, grafanaAlertNotification.DeepCopyObject(), grafanaAlertNotification.Status.GrafanaID) // todo: confirm this doesnt need null checking

	return &item
}
