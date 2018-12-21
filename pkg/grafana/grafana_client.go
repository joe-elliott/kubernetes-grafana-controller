package grafana

import (
	"errors"
	"net/http"
	"strings"

	"k8s.io/klog"
)

type Interface interface {
	PostDashboard(string) error
}

type GrafanaClient struct {
	address string
}

func NewGrafanaClient(address string) *GrafanaClient {

	client := &GrafanaClient{
		address: address,
	}

	return client
}

func (client *GrafanaClient) PostDashboard(dashboardJSON string) error {
	resp, err := http.Post(client.address+"/api/dashboards/db", "application/json", strings.NewReader(dashboardJSON))

	klog.Infof("http response: %v", resp)

	if resp != nil && resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	return err
}
