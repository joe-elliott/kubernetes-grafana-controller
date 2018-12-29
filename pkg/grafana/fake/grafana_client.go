package grafana

type GrafanaClientFake struct {
	address string
	fakeUID string

	PostedJson *string
}

func NewGrafanaClientFake(address string, fakeUID string) *GrafanaClientFake {

	client := &GrafanaClientFake{
		address:    address,
		fakeUID:    fakeUID,
		PostedJson: nil,
	}

	return client
}

func (client *GrafanaClientFake) PostDashboard(dashboardJSON string) (string, error) {
	client.PostedJson = &dashboardJSON

	return client.fakeUID, nil
}

func (client *GrafanaClientFake) DeleteDashboard(uid string) error {
	return nil
}
