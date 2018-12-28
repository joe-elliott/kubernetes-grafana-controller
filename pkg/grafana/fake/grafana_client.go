package grafana

type GrafanaClientFake struct {
	address string
	fakeUID string
}

func NewGrafanaClientFake(address string, fakeUID string) *GrafanaClientFake {

	client := &GrafanaClientFake{
		address: address,
		fakeUID: fakeUID,
	}

	return client
}

func (client *GrafanaClientFake) PostDashboard(dashboardJSON string) (string, error) {
	return client.fakeUID, nil
}

func (client *GrafanaClientFake) DeleteDashboard(uid string) error {
	return nil
}
