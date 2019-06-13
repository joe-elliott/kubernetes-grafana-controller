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

// FolderSyncer is the controller implementation for Folder resources
type FolderSyncer struct {
	grafanaFoldersLister listers.FolderLister
	grafanaClient        grafana.Interface
	grafanaclientset     clientset.Interface
}

// NewFolderController returns a new grafana Folder controller
func NewFolderController(
	grafanaclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaFolderInformer informers.FolderInformer) *Controller {

	syncer := &FolderSyncer{
		grafanaFoldersLister: grafanaFolderInformer.Lister(),
		grafanaClient:        grafanaClient,
		grafanaclientset:     grafanaclientset,
	}

	controller := NewController(grafanaFolderInformer.Informer(),
		kubeclientset,
		syncer)

	return controller
}

func (s *FolderSyncer) getType() string {
	return prometheus.TypeFolder
}

func (s *FolderSyncer) getRuntimeObjectByName(name string, namespace string) (runtime.Object, error) {
	return s.grafanaFoldersLister.Folders(namespace).Get(name)
}

func (s *FolderSyncer) deleteObjectById(id string) error {
	return s.grafanaClient.DeleteFolder(id)
}

func (s *FolderSyncer) updateObject(object runtime.Object) error {

	grafanaFolder, ok := object.(*v1alpha1.Folder)
	if !ok {
		return fmt.Errorf("expected folder in but got %#v", object)
	}

	id, idForDashboards, err := s.grafanaClient.PostFolder(grafanaFolder.Spec.JSON, grafanaFolder.Status.GrafanaID)

	if err != nil {
		return err
	}

	grafanaFolderCopy := grafanaFolder.DeepCopy()
	grafanaFolderCopy.Status.GrafanaID = id
	grafanaFolderCopy.Status.GrafanaIDForDashboards = idForDashboards

	_, err = s.grafanaclientset.GrafanaV1alpha1().Folders(grafanaFolder.Namespace).UpdateStatus(grafanaFolderCopy)
	if err != nil {
		return err
	}
	return nil
}

func (s *FolderSyncer) getAllKubernetesObjectIDs() ([]string, error) {
	Folders, err := s.grafanaFoldersLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)

	for _, Folder := range Folders {
		ids = append(ids, Folder.Status.GrafanaID)
	}

	return ids, nil
}

func (s *FolderSyncer) getAllGrafanaObjectIDs() ([]string, error) {
	return s.grafanaClient.GetAllFolderIds()
}

func (s *FolderSyncer) createWorkQueueItem(obj interface{}) *WorkQueueItem {
	var key string
	var err error
	var folder *v1alpha1.Folder
	var ok bool

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return nil
	}

	if folder, ok = obj.(*v1alpha1.Folder); !ok {
		utilruntime.HandleError(fmt.Errorf("expected folder in workqueue but got %#v", obj))
		return nil
	}

	item := NewWorkQueueItem(key, folder.DeepCopyObject(), folder.Status.GrafanaID) // todo: confirm this doesnt need null checking

	return &item
}
