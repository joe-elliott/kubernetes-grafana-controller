package grafana

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/imroc/req"
)

type Interface interface {
	PostDashboard(string) (string, error)
	DeleteDashboard(string) error

	PostNotificationChannel(string) (string, error)
	DeleteNotificationChannel(string) error
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
	//  todo:  decide and document how to handle having the unique name key alongside the id
	//         pass id into this function instead of querying grafana?

	var postChannel NotificationChannel
	err := json.Unmarshal([]byte(notificationChannelJson), &postChannel)

	if err != nil {
		return "", err
	}

	// Request existing notification channels
	resp, err := req.Get(client.address + "/api/alert-notifications")

	var responseBody []NotificationChannel

	err = resp.ToJSON(&responseBody)

	if err != nil {
		return "", err
	}

	var matchingChannel *NotificationChannel = nil

	for _, channel := range responseBody {
		if channel.Name == postChannel.Name {
			//found the thing
			matchingChannel = &channel
			break
		}
	}

	if matchingChannel != nil {

		if matchingChannel.ID == nil {
			return "", errors.New("Found a matching channel but id is nil")
		}

		// grafana requires an ID on put
		postChannel.ID = matchingChannel.ID

		postJSON, err := json.Marshal(postChannel)

		if err != nil {
			return "", err
		}

		return client.putGrafanaObject(string(postJSON), "/api/alert-notifications/"+strconv.Itoa(*matchingChannel.ID), "id")

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
