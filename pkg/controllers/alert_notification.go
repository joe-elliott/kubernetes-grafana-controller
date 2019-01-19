package controllers

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	"kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	clientset "kubernetes-grafana-controller/pkg/client/clientset/versioned"
	"kubernetes-grafana-controller/pkg/client/clientset/versioned/scheme"
	grafanascheme "kubernetes-grafana-controller/pkg/client/clientset/versioned/scheme"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions/grafana/v1alpha1"
	listers "kubernetes-grafana-controller/pkg/client/listers/grafana/v1alpha1"
	"kubernetes-grafana-controller/pkg/grafana"

	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// AlertNotificationSyncer is the controller implementation for GrafanaAlertNotification resources
type AlertNotificationSyncer struct {
	grafanaAlertNotificationLister listers.AlertNotificationLister
	grafanaClient                  grafana.Interface
	grafanaclientset               clientset.Interface
	recorder                       record.EventRecorder
}

// NewAlertNotificationController returns a new grafana alert notification controller
func NewAlertNotificationController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaAlertNotificationInformer informers.AlertNotificationInformer) *Controller {

	controllerAgentName := "grafana-alertnotification-controller"

	// Create event broadcaster
	// Add grafana-controller types to the default Kubernetes Scheme so Events can be
	// logged for grafana-controller types.
	utilruntime.Must(grafanascheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	syncer := &AlertNotificationSyncer{
		grafanaAlertNotificationLister: grafanaAlertNotificationInformer.Lister(),
		grafanaClient:                  grafanaClient,
		grafanaclientset:               grafanaclientset,
		recorder:                       recorder,
	}

	controller := NewController(controllerAgentName,
		grafanaAlertNotificationInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the GrafanaAlertNotification resource
// with the current status of the resource.
func (s *AlertNotificationSyncer) syncHandler(item WorkQueueItem) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(item.key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", item.key))
		return nil
	}

	// Get the AlertNotification resource with this namespace/name
	grafanaAlertNotification, err := s.grafanaAlertNotificationLister.AlertNotifications(namespace).Get(name)
	if err != nil {
		// The AlertNotification resource may no longer exist, in which case we stop
		// processing.
		if k8serrors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("grafanaAlertNotification '%s' in work queue no longer exists", item.key))

			// notification was deleted, so delete from grafana
			err = s.grafanaClient.DeleteAlertNotification(item.id)

			if err == nil {
				s.recorder.Event(item.originalObject, corev1.EventTypeNormal, SuccessDeleted, MessageResourceDeleted)
			}
		}

		return err
	}

	id, err := s.grafanaClient.PostAlertNotification(grafanaAlertNotification.Spec.JSON)

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the AlertNotification resource to reflect the
	// current state of the world
	err = s.updateGrafanaAlertNotificationStatus(grafanaAlertNotification, id)
	if err != nil {
		return err
	}

	s.recorder.Event(grafanaAlertNotification, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (s *AlertNotificationSyncer) updateGrafanaAlertNotificationStatus(grafanaAlertNotification *v1alpha1.AlertNotification, id string) error {

	if grafanaAlertNotification.Status.GrafanaID == id {
		return nil
	}

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	grafanaAlertNotificationCopy := grafanaAlertNotification.DeepCopy()
	grafanaAlertNotificationCopy.Status.GrafanaID = id
	// If the CustomResou	rceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the GrafanaAlertNotification resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.

	_, err := s.grafanaclientset.GrafanaV1alpha1().AlertNotifications(grafanaAlertNotification.Namespace).Update(grafanaAlertNotificationCopy)
	return err
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
