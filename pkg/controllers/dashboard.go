package controllers

import (
	"fmt"

	"errors"

	corev1 "k8s.io/api/core/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
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

// DashboardSyncer is the controller implementation for Dashboard resources
type DashboardSyncer struct {
	grafanaDashboardsLister listers.DashboardLister
	grafanaClient           grafana.Interface
	grafanaclientset        clientset.Interface
	recorder                record.EventRecorder
}

// NewDashboardController returns a new grafana dashboard controller
func NewDashboardController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaDashboardInformer informers.DashboardInformer) *Controller {

	controllerAgentName := "grafana-dashboard-controller"

	// Create event broadcaster
	// Add grafana-controller types to the default Kubernetes Scheme so Events can be
	// logged for grafana-controller types.
	utilruntime.Must(grafanascheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	syncer := &DashboardSyncer{
		grafanaDashboardsLister: grafanaDashboardInformer.Lister(),
		grafanaClient:           grafanaClient,
		grafanaclientset:        grafanaclientset,
		recorder:                recorder,
	}

	controller := NewController(controllerAgentName,
		grafanaDashboardInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Dashboard resource
// with the current status of the resource.
func (s *DashboardSyncer) syncHandler(item WorkQueueItem) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(item.key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", item.key))
		return nil
	}

	// Get the Dashboard resource with this namespace/name
	grafanaDashboard, err := s.grafanaDashboardsLister.Dashboards(namespace).Get(name)
	if err != nil {
		// The Dashboard resource may no longer exist, in which case we stop
		// processing.
		if k8serrors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("grafanaDashboard '%s' in work queue no longer exists", item.key))

			// dashboard was deleted, so delete from grafana
			err = s.grafanaClient.DeleteDashboard(item.id)

			if err == nil {
				s.recorder.Event(item.originalObject, corev1.EventTypeNormal, SuccessDeleted, MessageResourceDeleted)
			}
		}

		return err
	}

	id, err := s.grafanaClient.PostDashboard(grafanaDashboard.Spec.JSON)

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Dashboard resource to reflect the
	// current state of the world
	err = s.updateGrafanaDashboardStatus(grafanaDashboard, id)
	if err != nil {
		return err
	}

	s.recorder.Event(grafanaDashboard, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (s *DashboardSyncer) updateGrafanaDashboardStatus(grafanaDashboard *grafanav1alpha1.Dashboard, id string) error {

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	grafanaDashboardCopy := grafanaDashboard.DeepCopy()
	grafanaDashboardCopy.Status.GrafanaID = id
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the GrafanaDashboard resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.

	_, err := s.grafanaclientset.GrafanaV1alpha1().Dashboards(grafanaDashboard.Namespace).UpdateStatus(grafanaDashboardCopy)
	return err
}

func (s *DashboardSyncer) resyncDeletedObjects() error {

	// get all dashboards in grafana.  anything in grafana that's not in k8s gets nuked
	ids, err := s.grafanaClient.GetAllDashboardIds()

	if err != nil {
		return err
	}

	desiredDashboards, err := s.grafanaDashboardsLister.List(labels.Everything())

	if err != nil {
		return err
	}

	for _, id := range ids {
		var found = false

		for _, dashboard := range desiredDashboards {

			if dashboard.Status.GrafanaID == "" {
				return errors.New("found dashboard with unitialized state, bailing")
			}

			if dashboard.Status.GrafanaID == id {
				found = true
				break
			}
		}

		if !found {
			klog.Infof("Dashboard %v found in grafana but not k8s.  Deleting.", id)
			err = s.grafanaClient.DeleteDashboard(id)

			// if one fails just go ahead and bail out.  controlling logic will requeue
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *DashboardSyncer) createWorkQueueItem(obj interface{}) *WorkQueueItem {
	var key string
	var err error
	var dashboard *v1alpha1.Dashboard
	var ok bool

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return nil
	}

	if dashboard, ok = obj.(*v1alpha1.Dashboard); !ok {
		utilruntime.HandleError(fmt.Errorf("expected Dashboard in workqueue but got %#v", obj))
		return nil
	}

	item := NewWorkQueueItem(key, dashboard.DeepCopyObject(), dashboard.Status.GrafanaID) // todo: confirm this doesnt need null checking

	return &item
}
