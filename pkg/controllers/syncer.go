package controllers

type Syncer interface {
	syncHandler(item WorkQueueItem) error
	resyncDeletedObjects() error
	createWorkQueueItem(obj interface{}) *WorkQueueItem
}
