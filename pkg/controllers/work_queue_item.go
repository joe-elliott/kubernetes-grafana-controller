package controllers

import "k8s.io/apimachinery/pkg/runtime"

type WorkQueueItemType int

const (
	Update        = 1
	Delete        = 2
	ResyncDeleted = 3
)

type WorkQueueItem struct {
	itemType       WorkQueueItemType
	key            string
	originalObject runtime.Object
	id             string
}

func NewWorkQueueItem(itemType WorkQueueItemType, key string, originalObject runtime.Object, id string) WorkQueueItem {
	return WorkQueueItem{
		itemType:       itemType,
		key:            key,
		originalObject: originalObject,
		id:             id,
	}
}

func NewResyncDeletedObjects() WorkQueueItem {
	return WorkQueueItem{
		itemType:       ResyncDeleted,
		key:            "",
		originalObject: nil,
		id:             "",
	}
}

func (w *WorkQueueItem) isResyncDeletedObjects() bool {
	return w.itemType == ResyncDeleted
}
