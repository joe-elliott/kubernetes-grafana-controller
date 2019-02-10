package controllers

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

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

	controllerAgentName := "grafana-alertnotification-controller"

	syncer := &AlertNotificationSyncer{
		grafanaAlertNotificationLister: grafanaAlertNotificationInformer.Lister(),
		grafanaClient:                  grafanaClient,
		grafanaclientset:               grafanaclientset,
	}

	controller := NewController(controllerAgentName,
		grafanaAlertNotificationInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
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

func (s *AlertNotificationSyncer) resyncDeletedObjects() error {
	// get all alertNotification in grafana.  anything in grafana that's not in k8s gets nuked
	ids, err := s.grafanaClient.GetAllAlertNotificationIds()

	if err != nil {
		return err
	}

	desiredAlertNotifications, err := s.grafanaAlertNotificationLister.List(labels.Everything())

	if err != nil {
		return err
	}

	for _, id := range ids {
		var found = false

		for _, notification := range desiredAlertNotifications {

			if notification.Status.GrafanaID == "" {
				return errors.New("found notification with unitialized state, bailing")
			}

			if notification.Status.GrafanaID == id {
				found = true
				break
			}
		}

		if !found {
			klog.Infof("Notification %v found in grafana but not k8s.  Deleting.", id)
			err = s.grafanaClient.DeleteAlertNotification(id)

			// if one fails just go ahead and bail out.  controlling logic will requeue
			if err != nil {
				return err
			}
		}
	}

	return nil
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
