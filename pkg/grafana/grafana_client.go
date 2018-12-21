package grafana

import (
	"errors"
	"fmt"
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
	postJSON := fmt.Sprintf(`{
		"dashboard": %v,
		"folderId": 0,
		"overwrite": true
	}`, dashboardJSON)

	resp, err := http.Post(client.address+"/api/dashboards/db", "application/json", strings.NewReader(postJSON))

	klog.Infof("http response: %v", resp)

	if resp != nil && (resp.StatusCode >= 300 || resp.StatusCode < 200) {
		return errors.New(resp.Status)
	}

	return err
}
