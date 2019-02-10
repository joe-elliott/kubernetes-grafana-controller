package grafana

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/imroc/req"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const NO_ID = ""

type Interface interface {
	PostDashboard(string) (string, error)
	DeleteDashboard(string) error
	GetAllDashboardIds() ([]string, error)

	PostAlertNotification(string) (string, error)
	DeleteAlertNotification(string) error
	GetAllAlertNotificationIds() ([]string, error)

	PostDataSource(string, string) (string, error)
	DeleteDataSource(string) error
	GetAllDataSourceIds() ([]string, error)
}

type Client struct {
	address string
}

func NewClient(address string) *Client {

	client := &Client{
		address: address,
	}

	return client
}

func (client *Client) PostDashboard(dashboardJSON string) (string, error) {
	postJSON := fmt.Sprintf(`{
		"dashboard": %v,
		"folderId": 0,
		"overwrite": true
	}`, dashboardJSON)

	return client.postGrafanaObject(postJSON, "/api/dashboards/db", "uid")
}

func (client *Client) DeleteDashboard(id string) error {
	resp, err := req.Delete(client.address + "/api/dashboards/uid/" + id)

	if err != nil {
		return err
	}

	if !responseIsSuccess(resp) {
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

	if err = resp.ToJSON(&dashboards); err != nil {
		return nil, err
	}

	var ids []string

	for _, dashboard := range dashboards {
		ids = append(ids, dashboard["uid"].(string))
	}

	return ids, nil
}

func (client *Client) PostAlertNotification(alertNotificationJson string) (string, error) {

	// Grafana throws a 500 if you post 2 notification channels with the same name
	//  search for a matching notification channel and put these changes to it

	var postChannel map[string]interface{}
	err := json.Unmarshal([]byte(alertNotificationJson), &postChannel)

	if err != nil {
		return "", err
	}

	// Request existing notification channels
	resp, err := req.Get(client.address + "/api/alert-notifications")

	var responseBody []map[string]interface{}
	err = resp.ToJSON(&responseBody)

	if err != nil {
		return "", err
	}

	var matchingChannel *map[string]interface{}

	for _, channel := range responseBody {
		if channel["name"] == postChannel["name"] {
			//found the thing
			matchingChannel = &channel
			break
		}
	}

	if matchingChannel != nil {

		if (*matchingChannel)["id"] == nil {
			return "", errors.New("Found a matching channel but id is nil")
		}

		// grafana requires an ID on put
		postChannel["id"] = (*matchingChannel)["id"]
		postJSON, err := json.Marshal(postChannel)

		if err != nil {
			return "", err
		}

		return client.putGrafanaObject(string(postJSON), fmt.Sprintf("/api/alert-notifications/%v", postChannel["id"]), "id")

	} else {
		return client.postGrafanaObject(alertNotificationJson, "/api/alert-notifications", "id")
	}
}

func (client *Client) DeleteAlertNotification(id string) error {
	resp, err := req.Delete(client.address + "/api/alert-notifications/" + id)

	if err != nil {
		return err
	}

	if !responseIsSuccess(resp) {
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

	dataSourceJson, err := sanitizeObject(dataSourceJson)

	if err != nil {
		return "", err
	}

	if id != NO_ID {
		id, err := client.putGrafanaObject(dataSourceJson, fmt.Sprintf("/api/datasources/%v", id), "id")

		if err != nil {
			runtime.HandleError(err)

			return client.postGrafanaObject(dataSourceJson, "/api/datasources", "id")
		} else {
			return id, err
		}

	} else {
		return client.postGrafanaObject(dataSourceJson, "/api/datasources", "id")
	}
}

func (client *Client) DeleteDataSource(id string) error {
	resp, err := req.Delete(client.address + "/api/datasources/" + id)

	if err != nil {
		return err
	}

	if !responseIsSuccess(resp) {
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

	if err = resp.ToJSON(&datasources); err != nil {
		return nil, err
	}

	var ids []string

	for _, datasource := range datasources {
		ids = append(ids, fmt.Sprintf("%v", datasource["id"]))
	}

	return ids, nil
}

//
// shared
//

func (client *Client) postGrafanaObject(postJSON string, path string, idField string) (string, error) {
	var responseBody map[string]interface{}

	header := req.Header{
		"Content-Type": "application/json",
	}

	resp, err := req.Post(client.address+path, header, postJSON)

	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", errors.New("Error and response are nil")
	}

	if !responseIsSuccess(resp) {
		return "", errors.New(resp.Response().Status)
	}

	err = resp.ToJSON(&responseBody)

	if err != nil {
		return "", err
	}

	id, ok := responseBody[idField]

	if !ok {
		return "", fmt.Errorf("Response Body did not have field %s", idField)
	}

	// is there a better way to generically convert to string?
	idString := fmt.Sprintf("%v", id)

	return idString, nil
}

func (client *Client) putGrafanaObject(putJSON string, path string, idField string) (string, error) {
	var responseBody map[string]interface{}

	header := req.Header{
		"Content-Type": "application/json",
	}

	resp, err := req.Put(client.address+path, header, putJSON)

	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", errors.New("Error and response are nil")
	}

	if !responseIsSuccess(resp) {
		body, _ := resp.ToString()
		return "", errors.New(resp.Response().Status + ": " + body)
	}

	err = resp.ToJSON(&responseBody)

	if err != nil {
		return "", err
	}

	id, ok := responseBody[idField]

	if !ok {
		return "", fmt.Errorf("Response Body did not have field %s", idField)
	}

	// is there a better way to generically convert to string?
	idString := fmt.Sprintf("%v", id)

	return idString, nil
}

func responseIsSuccess(resp *req.Resp) bool {
	return resp.Response().StatusCode < 300 && resp.Response().StatusCode >= 200
}

func sanitizeObject(obj string) (string, error) {
	var jsonObject map[string]interface{}

	err := json.Unmarshal([]byte(obj), &jsonObject)
	if err != nil {
		return "", err
	}

	delete(jsonObject, "id")
	delete(jsonObject, "version")

	sanitizedBytes, err := json.Marshal(jsonObject)
	if err != nil {
		return "", err
	}

	return string(sanitizedBytes), nil
}
