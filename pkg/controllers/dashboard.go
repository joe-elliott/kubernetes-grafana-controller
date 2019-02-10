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

// DashboardSyncer is the controller implementation for Dashboard resources
type DashboardSyncer struct {
	grafanaDashboardsLister listers.DashboardLister
	grafanaClient           grafana.Interface
	grafanaclientset        clientset.Interface
}

// NewDashboardController returns a new grafana dashboard controller
func NewDashboardController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaDashboardInformer informers.DashboardInformer) *Controller {

	controllerAgentName := "grafana-dashboard-controller"

	syncer := &DashboardSyncer{
		grafanaDashboardsLister: grafanaDashboardInformer.Lister(),
		grafanaClient:           grafanaClient,
		grafanaclientset:        grafanaclientset,
	}

	controller := NewController(controllerAgentName,
		grafanaDashboardInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

func (s *DashboardSyncer) getRuntimeObjectByName(name string, namespace string) (runtime.Object, error) {
	return s.grafanaDashboardsLister.Dashboards(namespace).Get(name)
}

func (s *DashboardSyncer) deleteObjectById(id string) error {
	return s.grafanaClient.DeleteDashboard(id)
}

func (s *DashboardSyncer) updateObject(object runtime.Object) error {

	grafanaDashboard, ok := object.(*v1alpha1.Dashboard)
	if !ok {
		return fmt.Errorf("expected dashboard in but got %#v", object)
	}

	id, err := s.grafanaClient.PostDashboard(grafanaDashboard.Spec.JSON, grafanaDashboard.Status.GrafanaID)

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
