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

// DataSourceSyncer is the controller implementation for DataSource resources
type DataSourceSyncer struct {
	grafanaDataSourcesLister listers.DataSourceLister
	grafanaClient            grafana.Interface
	grafanaclientset         clientset.Interface
}

// NewDataSourceController returns a new grafana DataSource controller
func NewDataSourceController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaDataSourceInformer informers.DataSourceInformer) *Controller {

	controllerAgentName := "grafana-DataSource-controller"

	syncer := &DataSourceSyncer{
		grafanaDataSourcesLister: grafanaDataSourceInformer.Lister(),
		grafanaClient:            grafanaClient,
		grafanaclientset:         grafanaclientset,
	}

	controller := NewController(controllerAgentName,
		grafanaDataSourceInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

func (s *DataSourceSyncer) resyncDeletedObjects() error {
	// get all dashboards in grafana.  anything in grafana that's not in k8s gets nuked
	ids, err := s.grafanaClient.GetAllDataSourceIds()

	if err != nil {
		return err
	}

	desiredDataSources, err := s.grafanaDataSourcesLister.List(labels.Everything())

	if err != nil {
		return err
	}

	for _, id := range ids {
		var found = false

		for _, datasource := range desiredDataSources {

			if datasource.Status.GrafanaID == "" {
				return errors.New("found datasource with unitialized state, bailing")
			}

			if datasource.Status.GrafanaID == id {
				found = true
				break
			}
		}

		if !found {
			klog.Infof("Datasource %v found in grafana but not k8s.  Deleting.", id)
			err = s.grafanaClient.DeleteDataSource(id)

			// if one fails just go ahead and bail out.  controlling logic will requeue
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *DataSourceSyncer) getRuntimeObjectByName(name string, namespace string) (runtime.Object, error) {
	return s.grafanaDataSourcesLister.DataSources(namespace).Get(name)
}

func (s *DataSourceSyncer) deleteObjectById(id string) error {
	return s.grafanaClient.DeleteDataSource(id)
}

func (s *DataSourceSyncer) updateObject(object runtime.Object) error {

	grafanaDataSource, ok := object.(*v1alpha1.DataSource)
	if !ok {
		return fmt.Errorf("expected dataSource in but got %#v", object)
	}

	id, err := s.grafanaClient.PostDataSource(grafanaDataSource.Spec.JSON, grafanaDataSource.Status.GrafanaID)

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the GrafanaDataSource resource to reflect the
	// current state of the world
	grafanaDataSourceCopy := grafanaDataSource.DeepCopy()
	grafanaDataSourceCopy.Status.GrafanaID = id
	_, err = s.grafanaclientset.GrafanaV1alpha1().DataSources(grafanaDataSource.Namespace).UpdateStatus(grafanaDataSourceCopy)

	if err != nil {
		return err
	}
	return nil
}

func (s *DataSourceSyncer) createWorkQueueItem(obj interface{}) *WorkQueueItem {
	var key string
	var err error
	var dataSource *v1alpha1.DataSource
	var ok bool

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return nil
	}

	if dataSource, ok = obj.(*v1alpha1.DataSource); !ok {
		utilruntime.HandleError(fmt.Errorf("expected dataSource in workqueue but got %#v", obj))
		return nil
	}

	item := NewWorkQueueItem(key, dataSource.DeepCopyObject(), dataSource.Status.GrafanaID) // todo: confirm this doesnt need null checking

	return &item
}
