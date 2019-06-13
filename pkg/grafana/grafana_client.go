package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"kubernetes-grafana-controller/pkg/prometheus"

	"github.com/imroc/req"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const NO_ID = ""

type Interface interface {
	PostDashboard(string, string) (string, error)
	DeleteDashboard(string) error
	GetAllDashboardIds() ([]string, error)

	PostAlertNotification(string, string) (string, error)
	DeleteAlertNotification(string) error
	GetAllAlertNotificationIds() ([]string, error)

	PostDataSource(string, string) (string, error)
	DeleteDataSource(string) error
	GetAllDataSourceIds() ([]string, error)

	PostFolder(string, string) (string, error)
	DeleteFolder(string) error
	GetAllFolderIds() ([]string, error)
}

type Client struct {
	address string
}

func init() {
	// cost is required for prom metrics
	req.SetFlags(req.LstdFlags | req.Lcost)
}

func NewClient(address string) *Client {

	client := &Client{
		address: address,
	}

	return client
}

func (client *Client) PostDashboard(dashboardJSON string, uid string) (string, error) {
	dashboardJSON, err := sanitizeObject(dashboardJSON, false)

	if err != nil {
		return "", err
	}

	if uid != NO_ID {
		dashboardJSON, err = setId(dashboardJSON, "uid", uid)

		if err != nil {
			return "", err
		}
	}

	postJSON := fmt.Sprintf(`{
		"dashboard": %v,
		"folderId": 0,
		"overwrite": true
	}`, dashboardJSON)

	response, err := client.postGrafanaObject(postJSON, "/api/dashboards/db", prometheus.TypeDashboard)

	if err != nil {
		return "", err
	}

	return getField(response, "uid")
}

func (client *Client) DeleteDashboard(id string) error {
	resp, err := req.Delete(client.address + "/api/dashboards/uid/" + id)
	prometheus.GrafanaDeleteLatencyMilliseconds.WithLabelValues(prometheus.TypeDashboard).Observe(float64(resp.Cost() / time.Millisecond))

	if err != nil {
		return err
	}

	if !responseIsSuccessOrNotFound(resp) {
		return errors.New(resp.Response().Status)
	}

	return nil
}

func (client *Client) GetAllDashboardIds() ([]string, error) {
	var resp *req.Resp
	var err error
	var dashboards []map[string]interface{}

	// Request existing notification channels
	if resp, err = req.Get(client.address + "/api/search"); err != nil {
		return nil, err
	}
	prometheus.GrafanaGetLatencyMilliseconds.WithLabelValues(prometheus.TypeDashboard).Observe(float64(resp.Cost() / time.Millisecond))

	if err = resp.ToJSON(&dashboards); err != nil {
		return nil, err
	}

	var ids []string

	for _, dashboard := range dashboards {

		// folders will show up under the search api just like dashboards.  skip them or the resync logic will delete folders
		if dashboard["type"] == "dash-folder" {
			continue
		}

		ids = append(ids, dashboard["uid"].(string))
	}

	return ids, nil
}

func (client *Client) PostAlertNotification(alertNotificationJson string, id string) (string, error) {
	var response map[string]interface{}
	alertNotificationJson, err := sanitizeObject(alertNotificationJson, false)

	if err != nil {
		return "", err
	}

	if id == NO_ID {
		response, err = client.postGrafanaObject(alertNotificationJson, "/api/alert-notifications", prometheus.TypeAlertNotification)

		if err != nil {
			return "", err
		}
	} else {
		// alert notification requires the id in the object for unknown reasons
		alertNotificationJson, err = setId(alertNotificationJson, "id", id)

		if err != nil {
			return "", err
		}

		response, err = client.putGrafanaObject(alertNotificationJson, fmt.Sprintf("/api/alert-notifications/%v", id), prometheus.TypeAlertNotification)

		// try a put if the post fails
		if err != nil {
			runtime.HandleError(err)
			prometheus.GrafanaWastedPutTotal.WithLabelValues(prometheus.TypeAlertNotification).Inc()

			response, err = client.postGrafanaObject(alertNotificationJson, "/api/alert-notifications", prometheus.TypeAlertNotification)

			if err != nil {
				return "", err
			}
		}
	}

	return getField(response, "id")
}

func (client *Client) DeleteAlertNotification(id string) error {
	resp, err := req.Delete(client.address + "/api/alert-notifications/" + id)
	prometheus.GrafanaDeleteLatencyMilliseconds.WithLabelValues(prometheus.TypeAlertNotification).Observe(float64(resp.Cost() / time.Millisecond))

	if err != nil {
		return err
	}

	if !responseIsSuccessOrNotFound(resp) {
		return errors.New(resp.Response().Status)
	}

	return nil
}

func (client *Client) GetAllAlertNotificationIds() ([]string, error) {
	var resp *req.Resp
	var err error
	var channels []map[string]interface{}

	// Request existing notification channels
	if resp, err = req.Get(client.address + "/api/alert-notifications"); err != nil {
		return nil, err
	}
	prometheus.GrafanaGetLatencyMilliseconds.WithLabelValues(prometheus.TypeAlertNotification).Observe(float64(resp.Cost() / time.Millisecond))

	if err = resp.ToJSON(&channels); err != nil {
		return nil, err
	}

	var ids []string

	for _, channel := range channels {
		ids = append(ids, fmt.Sprintf("%v", channel["id"]))
	}

	return ids, nil
}

func (client *Client) PostDataSource(dataSourceJson string, id string) (string, error) {
	var response map[string]interface{}
	dataSourceJson, err := sanitizeObject(dataSourceJson, false)

	if err != nil {
		return "", err
	}

	if id == NO_ID {
		response, err = client.postGrafanaObject(dataSourceJson, "/api/datasources", prometheus.TypeDataSource)

		if err != nil {
			return "", err
		}
	} else {
		response, err = client.putGrafanaObject(dataSourceJson, fmt.Sprintf("/api/datasources/%v", id), prometheus.TypeDataSource)

		if err != nil {
			runtime.HandleError(err)
			prometheus.GrafanaWastedPutTotal.WithLabelValues(prometheus.TypeDataSource).Inc()

			response, err = client.postGrafanaObject(dataSourceJson, "/api/datasources", prometheus.TypeDataSource)

			if err != nil {
				return "", err
			}
		}
	}

	return getField(response, "id")
}

func (client *Client) DeleteDataSource(id string) error {
	resp, err := req.Delete(client.address + "/api/datasources/" + id)
	prometheus.GrafanaDeleteLatencyMilliseconds.WithLabelValues(prometheus.TypeDataSource).Observe(float64(resp.Cost() / time.Millisecond))

	if err != nil {
		return err
	}

	if !responseIsSuccessOrNotFound(resp) {
		return errors.New(resp.Response().Status)
	}

	return nil
}

func (client *Client) GetAllDataSourceIds() ([]string, error) {
	var resp *req.Resp
	var err error
	var datasources []map[string]interface{}

	// Request existing notification channels
	if resp, err = req.Get(client.address + "/api/datasources"); err != nil {
		return nil, err
	}
	prometheus.GrafanaGetLatencyMilliseconds.WithLabelValues(prometheus.TypeDataSource).Observe(float64(resp.Cost() / time.Millisecond))

	if err = resp.ToJSON(&datasources); err != nil {
		return nil, err
	}

	var ids []string

	for _, datasource := range datasources {
		ids = append(ids, fmt.Sprintf("%v", datasource["id"]))
	}

	return ids, nil
}

func (client *Client) PostFolder(folderJson string, id string) (string, error) {
	var response map[string]interface{}
	folderJson, err := sanitizeObject(folderJson, true)

	if err != nil {
		return "", err
	}

	if id == NO_ID {
		response, err = client.postGrafanaObject(folderJson, "/api/folders", prometheus.TypeFolder)

		if err != nil {
			return "", err
		}
	} else {
		response, err = client.putGrafanaObject(folderJson, fmt.Sprintf("/api/folders/%v", id), prometheus.TypeFolder)

		if err != nil {
			runtime.HandleError(err)
			prometheus.GrafanaWastedPutTotal.WithLabelValues(prometheus.TypeFolder).Inc()

			response, err = client.postGrafanaObject(folderJson, "/api/folders", prometheus.TypeFolder)

			if err != nil {
				return "", err
			}
		}
	}

	return getField(response, "uid")
}

func (client *Client) DeleteFolder(id string) error {
	resp, err := req.Delete(client.address + "/api/folders/" + id)
	prometheus.GrafanaDeleteLatencyMilliseconds.WithLabelValues(prometheus.TypeFolder).Observe(float64(resp.Cost() / time.Millisecond))

	if err != nil {
		return err
	}

	if !responseIsSuccessOrNotFound(resp) {
		return errors.New(resp.Response().Status)
	}

	return nil
}

func (client *Client) GetAllFolderIds() ([]string, error) {
	var resp *req.Resp
	var err error
	var folders []map[string]interface{}

	// Request existing notification channels
	if resp, err = req.Get(client.address + "/api/folders"); err != nil {
		return nil, err
	}
	prometheus.GrafanaGetLatencyMilliseconds.WithLabelValues(prometheus.TypeFolder).Observe(float64(resp.Cost() / time.Millisecond))

	if err = resp.ToJSON(&folders); err != nil {
		return nil, err
	}

	var ids []string

	for _, folder := range folders {
		ids = append(ids, fmt.Sprintf("%v", folder["uid"]))
	}

	return ids, nil
}

//
// shared
//

func (client *Client) postGrafanaObject(postJSON string, path string, prometheusType string) (map[string]interface{}, error) {
	var responseBody map[string]interface{}

	header := req.Header{
		"Content-Type": "application/json",
	}

	resp, err := req.Post(client.address+path, header, postJSON)
	prometheus.GrafanaPostLatencyMilliseconds.WithLabelValues(prometheusType).Observe(float64(resp.Cost() / time.Millisecond))

	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("Error and response are nil")
	}

	if !responseIsSuccess(resp) {
		return nil, errors.New(resp.Response().Status)
	}

	err = resp.ToJSON(&responseBody)

	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (client *Client) putGrafanaObject(putJSON string, path string, prometheusType string) (map[string]interface{}, error) {
	var responseBody map[string]interface{}

	header := req.Header{
		"Content-Type": "application/json",
	}

	resp, err := req.Put(client.address+path, header, putJSON)
	prometheus.GrafanaPutLatencyMilliseconds.WithLabelValues(prometheusType).Observe(float64(resp.Cost() / time.Millisecond))

	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("Error and response are nil")
	}

	if !responseIsSuccess(resp) {
		body, _ := resp.ToString()
		return nil, errors.New(resp.Response().Status + ": " + body)
	}

	err = resp.ToJSON(&responseBody)

	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func responseIsSuccess(resp *req.Resp) bool {
	return resp.Response().StatusCode < 300 && resp.Response().StatusCode >= 200
}

func responseIsSuccessOrNotFound(resp *req.Resp) bool {
	return responseIsSuccess(resp) || resp.Response().StatusCode == 404
}

func sanitizeObject(obj string, addOverwrite bool) (string, error) {
	var jsonObject map[string]interface{}

	err := json.Unmarshal([]byte(obj), &jsonObject)
	if err != nil {
		return "", err
	}

	delete(jsonObject, "id")
	delete(jsonObject, "version")

	// some grafana apis require overwrite = true to ignore versions
	if addOverwrite {
		jsonObject["overwrite"] = true
	}

	sanitizedBytes, err := json.Marshal(jsonObject)
	if err != nil {
		return "", err
	}

	return string(sanitizedBytes), nil
}

func setId(obj string, idField string, idValue string) (string, error) {
	var jsonObject map[string]interface{}

	err := json.Unmarshal([]byte(obj), &jsonObject)
	if err != nil {
		return "", err
	}

	intValue, err := strconv.Atoi(idValue)
	if err == nil {
		jsonObject[idField] = intValue
	} else {
		jsonObject[idField] = idValue
	}

	bytes, err := json.Marshal(jsonObject)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func getField(obj map[string]interface{}, field string) (string, error) {

	id, ok := obj[field]

	if !ok {
		return "", fmt.Errorf("Map did not have field %s", field)
	}

	// is there a better way to generically convert to string?
	idString := fmt.Sprintf("%v", id)

	return idString, nil
}
