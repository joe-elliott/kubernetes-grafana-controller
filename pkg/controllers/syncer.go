package controllers

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type Syncer interface {
	resyncDeletedObjects() error
	createWorkQueueItem(obj interface{}) *WorkQueueItem

	getRuntimeObjectByName(name string, namespace string) (runtime.Object, error)
	deleteObjectById(id string) error
	updateObject(object runtime.Object) error
}
