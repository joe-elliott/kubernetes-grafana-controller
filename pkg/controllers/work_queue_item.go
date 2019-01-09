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
