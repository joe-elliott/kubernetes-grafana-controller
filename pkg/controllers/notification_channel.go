package controllers

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	"kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	grafanav1alpha1 "kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	clientset "kubernetes-grafana-controller/pkg/client/clientset/versioned"
	"kubernetes-grafana-controller/pkg/client/clientset/versioned/scheme"
	grafanascheme "kubernetes-grafana-controller/pkg/client/clientset/versioned/scheme"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions/grafana/v1alpha1"
	listers "kubernetes-grafana-controller/pkg/client/listers/grafana/v1alpha1"
	"kubernetes-grafana-controller/pkg/grafana"

	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NotificationChannelSyncer is the controller implementation for GrafanaNotificationChannel resources
type NotificationChannelSyncer struct {
	grafanaNotificationChannelLister listers.GrafanaNotificationChannelLister
	grafanaClient                    grafana.Interface
	grafanaclientset                 clientset.Interface
	recorder                         record.EventRecorder
}

// NewNotificationChannelController returns a new grafana channel controller
func NewNotificationChannelController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaNotificationChannelInformer informers.GrafanaNotificationChannelInformer) *Controller {

	controllerAgentName := "grafana-notificationchannel-controller"

	// Create event broadcaster
	// Add grafana-controller types to the default Kubernetes Scheme so Events can be
	// logged for grafana-controller types.
	utilruntime.Must(grafanascheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	syncer := &NotificationChannelSyncer{
		grafanaNotificationChannelLister: grafanaNotificationChannelInformer.Lister(),
		grafanaClient:                    grafanaClient,
		grafanaclientset:                 grafanaclientset,
		recorder:                         recorder,
	}

	controller := NewController(controllerAgentName,
		grafanaNotificationChannelInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the GrafanaNotificationChannel resource
// with the current status of the resource.
func (s *NotificationChannelSyncer) syncHandler(item WorkQueueItem) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(item.key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", item.key))
		return nil
	}

	// Get the GrafanaNotificationChannel resource with this namespace/name
	grafanaNotificationChannel, err := s.grafanaNotificationChannelLister.GrafanaNotificationChannels(namespace).Get(name)
	if err != nil {
		// The GrafanaNotificationChannel resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("grafanaNotificationChannel '%s' in work queue no longer exists", item.key))

			// channel was deleted, so delete from grafana
			err = s.grafanaClient.DeleteNotificationChannel(item.uuid)

			if err == nil {
				s.recorder.Event(item.originalObject, corev1.EventTypeNormal, SuccessDeleted, MessageResourceDeleted)
			}
		}

		return err
	}

	id, err := s.grafanaClient.PostNotificationChannel(grafanaNotificationChannel.Spec.NotificationChannelJSON)

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the GrafanaNotificationChannel resource to reflect the
	// current state of the world
	err = s.updateGrafanaNotificationChannelStatus(grafanaNotificationChannel, id)
	if err != nil {
		return err
	}

	s.recorder.Event(grafanaNotificationChannel, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (s *NotificationChannelSyncer) updateGrafanaNotificationChannelStatus(grafanaNotificationChannel *grafanav1alpha1.GrafanaNotificationChannel, id string) error {

	if grafanaNotificationChannel.Status.GrafanaID == id {
		return nil
	}

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	grafanaNotificationChannelCopy := grafanaNotificationChannel.DeepCopy()
	grafanaNotificationChannelCopy.Status.GrafanaID = id
	// If the CustomResou	rceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the GrafanaNotificationChannel resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.

	_, err := s.grafanaclientset.GrafanaV1alpha1().GrafanaNotificationChannels(grafanaNotificationChannel.Namespace).Update(grafanaNotificationChannelCopy)
	return err
}

func (s *NotificationChannelSyncer) enqueueResyncDeletedObjects() error {
	fmt.Println("resyncing all notification channels!")
	return nil
}

func (s *NotificationChannelSyncer) createWorkQueueItem(obj interface{}) *WorkQueueItem {
	var key string
	var err error
	var grafanaNotificationChannel *v1alpha1.GrafanaNotificationChannel
	var ok bool

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return nil
	}

	if grafanaNotificationChannel, ok = obj.(*v1alpha1.GrafanaNotificationChannel); !ok {
		utilruntime.HandleError(fmt.Errorf("expected GrafanaNotificationChannel in workqueue but got %#v", obj))
		return nil
	}

	item := NewWorkQueueItem(key, grafanaNotificationChannel.DeepCopyObject(), grafanaNotificationChannel.Status.GrafanaID) // todo: confirm this doesnt need null checking

	return &item
}
