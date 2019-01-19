package controllers

import "k8s.io/apimachinery/pkg/runtime"

type WorkQueueItem struct {
	key            string
	originalObject runtime.Object
	id             string
}

func NewWorkQueueItem(key string, originalObject runtime.Object, id string) WorkQueueItem {
	return WorkQueueItem{
		key:            key,
		originalObject: originalObject,
		id:             id,
	}
}

func NewResyncDeletedObjects() WorkQueueItem {
	return WorkQueueItem{
		key:            "",
		originalObject: nil,
		id:             "",
	}
}

func (w *WorkQueueItem) isResyncDeletedObjects() bool {
	return w.key == "" && w.originalObject == nil && w.id == ""
}
