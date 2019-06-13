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

func (client *ClientFake) PostDashboard(json string, uid string) (string, error) {
	client.PostedJson = &json

	return client.fakeID, nil
}

func (client *ClientFake) PostDashboardWithFolder(json string, folderId string, uid string) (string, error) {
	client.PostedJson = &json

	return client.fakeID, nil
}

func (client *ClientFake) DeleteDashboard(id string) error {
	return nil
}

func (client *ClientFake) GetAllDashboardIds() ([]string, error) {
	return nil, nil
}

func (client *ClientFake) PostAlertNotification(json string, id string) (string, error) {
	client.PostedJson = &json

	return client.fakeID, nil
}

func (client *ClientFake) DeleteAlertNotification(id string) error {
	return nil
}

func (client *ClientFake) PostDataSource(json string, id string) (string, error) {
	client.PostedJson = &json

	return client.fakeID, nil
}

func (client *ClientFake) DeleteDataSource(id string) error {
	return nil
}

func (client *ClientFake) GetAllDatasourceIds() ([]string, error) {
	return nil, nil
}

func (client *ClientFake) GetAllAlertNotificationIds() ([]string, error) {
	return nil, nil
}

func (client *ClientFake) PostFolder(json string, id string) (string, string, error) {
	client.PostedJson = &json

	return client.fakeID, 0, nil
}

func (client *ClientFake) DeleteFolder(id string) error {
	return nil
}

func (client *ClientFake) GetAllFolderIds() ([]string, error) {
	return nil, nil
}
