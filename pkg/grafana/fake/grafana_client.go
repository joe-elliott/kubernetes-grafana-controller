package grafana

type ClientFake struct {
	address string
	fakeID  string

	PostedJson *string
}

func NewGrafanaClientFake(address string, fakeID string) *ClientFake {

	client := &ClientFake{
		address:    address,
		fakeID:     fakeID,
		PostedJson: nil,
	}

	return client
}

func (client *ClientFake) PostDashboard(json string) (string, error) {
	client.PostedJson = &json

	return client.fakeID, nil
}

func (client *ClientFake) DeleteDashboard(id string) error {
	return nil
}

func (client *ClientFake) GetAllDashboardIds() ([]string, error) {
	return nil, nil
}

func (client *ClientFake) PostNotificationChannel(json string) (string, error) {
	client.PostedJson = &json

	return client.fakeID, nil
}

func (client *ClientFake) DeleteNotificationChannel(id string) error {
	return nil
}

func (client *ClientFake) PostDataSource(json string) (string, error) {
	client.PostedJson = &json

	return client.fakeID, nil
}

func (client *ClientFake) DeleteDataSource(id string) error {
	return nil
}

func (client *ClientFake) GetAllDatasourceIds() ([]string, error) {
	return nil, nil
}

func (client *ClientFake) GetAllNotificationChannelIds() ([]string, error) {
	return nil, nil
}
