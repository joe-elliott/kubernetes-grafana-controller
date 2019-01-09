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

// DataSourceSyncer is the controller implementation for GrafanaDataSource resources
type DataSourceSyncer struct {
	grafanaDataSourcesLister listers.GrafanaDataSourceLister
	grafanaClient            grafana.Interface
	grafanaclientset         clientset.Interface
	recorder                 record.EventRecorder
}

// NewDataSourceController returns a new grafana DataSource controller
func NewDataSourceController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaDataSourceInformer informers.GrafanaDataSourceInformer) *Controller {

	controllerAgentName := "grafana-DataSource-controller"

	// Create event broadcaster
	// Add grafana-controller types to the default Kubernetes Scheme so Events can be
	// logged for grafana-controller types.
	utilruntime.Must(grafanascheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	syncer := &DataSourceSyncer{
		grafanaDataSourcesLister: grafanaDataSourceInformer.Lister(),
		grafanaClient:            grafanaClient,
		grafanaclientset:         grafanaclientset,
		recorder:                 recorder,
	}

	controller := NewController(controllerAgentName,
		grafanaDataSourceInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the GrafanaDataSource resource
// with the current status of the resource.
func (s *DataSourceSyncer) syncHandler(item WorkQueueItem) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(item.key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", item.key))
		return nil
	}

	// Get the GrafanaDataSource resource with this namespace/name
	grafanaDataSource, err := s.grafanaDataSourcesLister.GrafanaDataSources(namespace).Get(name)
	if err != nil {
		// The GrafanaDataSource resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("grafanaDataSource '%s' in work queue no longer exists", item.key))

			// DataSource was deleted, so delete from grafana
			err = s.grafanaClient.DeleteDataSource(item.uuid)

			if err != nil {
				s.recorder.Event(grafanaDataSource, corev1.EventTypeNormal, SuccessDeleted, MessageResourceDeleted)
			}
		}

		return err
	}

	id, err := s.grafanaClient.PostDataSource(grafanaDataSource.Spec.DataSourceJSON)

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the GrafanaDataSource resource to reflect the
	// current state of the world
	err = s.updateGrafanaDataSourceStatus(grafanaDataSource, id)
	if err != nil {
		return err
	}

	s.recorder.Event(grafanaDataSource, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (s *DataSourceSyncer) updateGrafanaDataSourceStatus(grafanaDataSource *grafanav1alpha1.GrafanaDataSource, id string) error {

	if grafanaDataSource.Status.GrafanaID == id {
		return nil
	}

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	grafanaDataSourceCopy := grafanaDataSource.DeepCopy()
	grafanaDataSourceCopy.Status.GrafanaID = id
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the GrafanaDataSource resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.

	_, err := s.grafanaclientset.GrafanaV1alpha1().GrafanaDataSources(grafanaDataSource.Namespace).Update(grafanaDataSourceCopy)
	return err
}

func (s *DataSourceSyncer) createWorkQueueItem(obj interface{}) *WorkQueueItem {
	var key string
	var err error
	var dataSource *v1alpha1.GrafanaDataSource
	var ok bool

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return nil
	}

	if dataSource, ok = obj.(*v1alpha1.GrafanaDataSource); !ok {
		utilruntime.HandleError(fmt.Errorf("expected GrafanaDataSource in workqueue but got %#v", obj))
		return nil
	}

	item := NewWorkQueueItem(key, DataSource, dataSource.Status.GrafanaID) // todo: confirm this doesnt need null checking

	return &item
}
