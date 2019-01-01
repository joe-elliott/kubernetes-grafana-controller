package grafana

type NotificationChannel struct {
	ID           *int        `json:"id,omitempty"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	IsDefault    *bool       `json:"isDefault,omitempty"`
	SendReminder *bool       `json:"sendReminder,omitempty"`
	Frequency    *string     `json:"frequency,omitempty"`
	Settings     interface{} `json:"settings"`
}
