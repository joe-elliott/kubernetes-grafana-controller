package controllers

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type Syncer interface {
	createWorkQueueItem(obj interface{}) *WorkQueueItem
	deleteObjectById(id string) error

	// support basic sync handling
	getRuntimeObjectByName(name string, namespace string) (runtime.Object, error)
	updateObject(object runtime.Object) error

	// support deleted objects resync
	getAllKubernetesObjectIDs() ([]string, error)
	getAllGrafanaObjectIDs() ([]string, error)
}
