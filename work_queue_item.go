package main

type GrafanaObject int

const (
	Dashboard = 0
)

type WorkQueueItem struct {
	key        string
	objectType GrafanaObject
	uuid       string
}

func NewWorkQueueItem(key string, objectType GrafanaObject, uuid string) WorkQueueItem {
	return WorkQueueItem{
		key:        key,
		objectType: objectType,
		uuid:       uuid,
	}
}
