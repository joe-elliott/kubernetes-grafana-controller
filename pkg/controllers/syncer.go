package controllers

type Syncer interface {
	syncHandler(item WorkQueueItem) error
	resyncAll() error
	createWorkQueueItem(obj interface{}) *WorkQueueItem
}
