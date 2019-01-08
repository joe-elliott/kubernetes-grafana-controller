package grafana

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/imroc/req"
)

type Interface interface {
	PostDashboard(string) (string, error)
	DeleteDashboard(string) error

	PostNotificationChannel(string) (string, error)
	DeleteNotificationChannel(string) error

	PostDataSource(string) (string, error)
	DeleteDataSource(string) error
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

func (client *Client) DeleteDashboard(uid string) error {
	resp, err := req.Delete(client.address + "/api/dashboards/uid/" + uid)

	if err != nil {
		return err
	}

	if !responseIsSuccess(resp) {
		return errors.New(resp.Response().Status)
	}

	return nil
}

func (client *Client) PostNotificationChannel(notificationChannelJson string) (string, error) {

	// Grafana throws a 500 if you post 2 notification channels with the same name
	//  search for a matching notification channel and put these changes to it

	var postChannel map[string]interface{}
	err := json.Unmarshal([]byte(notificationChannelJson), &postChannel)

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
		return client.postGrafanaObject(notificationChannelJson, "/api/alert-notifications", "id")
	}
}

func (client *Client) DeleteNotificationChannel(id string) error {
	resp, err := req.Delete(client.address + "/api/alert-notifications/" + id)

	if err != nil {
		return err
	}

	if !responseIsSuccess(resp) {
		return errors.New(resp.Response().Status)
	}

	return nil
}

func (client *Client) PostDataSource(dataSourceJson string) (string, error) {

	// Grafana throws a 500 if you post 2 datasources with the same name
	//  search for a matching datasources and put these changes to it

	var postDataSource map[string]interface{}
	err := json.Unmarshal([]byte(dataSourceJson), &postDataSource)

	if err != nil {
		return "", err
	}

	// Request existing notification channels
	resp, err := req.Get(client.address + "/api/datasources")

	var responseBody []map[string]interface{}
	err = resp.ToJSON(&responseBody)

	if err != nil {
		return "", err
	}

	var matchingDataSource *map[string]interface{}

	for _, channel := range responseBody {
		if channel["name"] == postDataSource["name"] {
			//found the thing
			matchingDataSource = &channel
			break
		}
	}

	if matchingDataSource != nil {

		if (*matchingDataSource)["id"] == nil {
			return "", errors.New("Found a matching datasource but id is nil")
		}

		// grafana requires an ID on put
		postDataSource["id"] = (*matchingDataSource)["id"]
		postJSON, err := json.Marshal(postDataSource)

		if err != nil {
			return "", err
		}

		return client.putGrafanaObject(string(postJSON), fmt.Sprintf("/api/datasources/%v", postDataSource["id"]), "id")

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
		return "", fmt.Errorf("Respone Body did not have field %s", idField)
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
		return "", fmt.Errorf("Respone Body did not have field %s", idField)
	}

	// is there a better way to generically convert to string?
	idString := fmt.Sprintf("%v", id)

	return idString, nil
}

func responseIsSuccess(resp *req.Resp) bool {
	return resp.Response().StatusCode < 300 && resp.Response().StatusCode >= 200
}
