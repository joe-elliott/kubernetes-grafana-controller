package controllers

type GrafanaObject int

const (
	Dashboard           = 0
	NotificationChannel = 1
	DataSource          = 2
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
