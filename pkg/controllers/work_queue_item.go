package controllers

import "k8s.io/apimachinery/pkg/runtime"

type WorkQueueItem struct {
	key            string
	originalObject runtime.Object
	uuid           string
}

func NewWorkQueueItem(key string, originalObject runtime.Object, uuid string) WorkQueueItem {
	return WorkQueueItem{
		key:            key,
		originalObject: originalObject,
		uuid:           uuid,
	}
}

func NewResyncDeletedObjects() WorkQueueItem {
	return WorkQueueItem{
		key:            "",
		originalObject: nil,
		uuid:           "",
	}
}

func (w *WorkQueueItem) isResyncDeletedObjects() bool {
	return w.key == "" && w.originalObject == nil && w.uuid == ""
}
