package controllers

type Syncer interface {
	syncHandler(item WorkQueueItem) error
	enqueueResyncDeletedObjects() error
	createWorkQueueItem(obj interface{}) *WorkQueueItem
}
