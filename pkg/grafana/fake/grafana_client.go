package grafana

type ClientFake struct {
	address string
	fakeUID string

	PostedJson *string
}

func NewGrafanaClientFake(address string, fakeUID string) *ClientFake {

	client := &ClientFake{
		address:    address,
		fakeUID:    fakeUID,
		PostedJson: nil,
	}

	return client
}

func (client *ClientFake) PostDashboard(dashboardJSON string) (string, error) {
	client.PostedJson = &dashboardJSON

	return client.fakeUID, nil
}

func (client *ClientFake) DeleteDashboard(uid string) error {
	return nil
}
