package controllers

import (
	"fmt"
	"github.com/joe-elliott/kubernetes-grafana-controller/pkg/prometheus"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/joe-elliott/kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	clientset "github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/clientset/versioned"
	informers "github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/informers/externalversions/grafana/v1alpha1"
	listers "github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/listers/grafana/v1alpha1"
	"github.com/joe-elliott/kubernetes-grafana-controller/pkg/grafana"
)

// DashboardSyncer is the controller implementation for Dashboard resources
type DashboardSyncer struct {
	grafanaDashboardsLister listers.DashboardLister
	grafanaFoldersLister    listers.FolderLister
	grafanaClient           grafana.Interface
	grafanaclientset        clientset.Interface
}

// NewDashboardController returns a new grafana dashboard controller
func NewDashboardController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaDashboardInformer informers.DashboardInformer,
	grafanaFolderInformer informers.FolderInformer) *Controller {

	syncer := &DashboardSyncer{
		grafanaDashboardsLister: grafanaDashboardInformer.Lister(),
		grafanaFoldersLister:    grafanaFolderInformer.Lister(),
		grafanaClient:           grafanaClient,
		grafanaclientset:        grafanaclientset,
	}

	controller := NewController(grafanaDashboardInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

func (s *DashboardSyncer) getType() string {
	return prometheus.TypeDashboard
}

func (s *DashboardSyncer) getRuntimeObjectByName(name string, namespace string) (runtime.Object, error) {
	return s.grafanaDashboardsLister.Dashboards(namespace).Get(name)
}

func (s *DashboardSyncer) deleteObjectById(id string) error {
	return s.grafanaClient.DeleteDashboard(id)
}

func (s *DashboardSyncer) updateObject(object runtime.Object) error {
	var err error
	var id string

	grafanaDashboard, ok := object.(*v1alpha1.Dashboard)
	if !ok {
		return fmt.Errorf("expected dashboard in but got %#v", object)
	}

	if grafanaDashboard.Spec.FolderName != "" {
		folder, err := s.grafanaFoldersLister.Folders(grafanaDashboard.Namespace).Get(grafanaDashboard.Spec.FolderName)

		if err != nil {
			return err
		}

		id, err = s.grafanaClient.PostDashboardWithFolder(grafanaDashboard.Spec.JSON, folder.Status.GrafanaIDForDashboards, grafanaDashboard.Status.GrafanaID)
	} else {
		id, err = s.grafanaClient.PostDashboard(grafanaDashboard.Spec.JSON, grafanaDashboard.Status.GrafanaID)
	}

	if err != nil {
		return err
	}

	grafanaDashboardCopy := grafanaDashboard.DeepCopy()
	grafanaDashboardCopy.Status.GrafanaID = id

	_, err = s.grafanaclientset.GrafanaV1alpha1().Dashboards(grafanaDashboard.Namespace).UpdateStatus(grafanaDashboardCopy)
	if err != nil {
		return err
	}
	return nil
}

func (s *DashboardSyncer) getAllKubernetesObjectIDs() ([]string, error) {
	dashboards, err := s.grafanaDashboardsLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)

	for _, dashboard := range dashboards {
		ids = append(ids, dashboard.Status.GrafanaID)
	}

	return ids, nil
}

func (s *DashboardSyncer) getAllGrafanaObjectIDs() ([]string, error) {
	return s.grafanaClient.GetAllDashboardIds()
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
