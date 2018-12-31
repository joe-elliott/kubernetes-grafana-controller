package controllers

type Syncer interface {
	syncHandler(item WorkQueueItem) error
	createWorkQueueItem(obj interface{}) *WorkQueueItem
}
