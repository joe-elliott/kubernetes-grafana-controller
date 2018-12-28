package grafana

type GrafanaClientFake struct {
	address string
}

func NewGrafanaClientFake(address string) *GrafanaClientFake {

	client := &GrafanaClientFake{
		address: address,
	}

	return client
}

func (client *GrafanaClientFake) PostDashboard(dashboardJSON string) (string, error) {
	return "fakeUidString", nil
}

func (client *GrafanaClientFake) DeleteDashboard(uid string) error {
	return nil
}
