package grafana

import (
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
	var responseBody map[string]interface{}

	postJSON := fmt.Sprintf(`{
		"dashboard": %v,
		"folderId": 0,
		"overwrite": true
	}`, dashboardJSON)

	header := req.Header{
		"Content-Type": "application/json",
	}

	resp, err := req.Post(client.address+"/api/dashboards/db", header, postJSON)

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

	uid, ok := responseBody["uid"]

	if !ok {
		return "", errors.New("Response Body did not have uid")
	}

	uidString, ok := uid.(string)

	if !ok {
		return "", fmt.Errorf("Unable to convert uid %#v to string", uid)
	}

	return uidString, err
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
	var responseBody map[string]interface{}

	postJSON := notificationChannelJson
	header := req.Header{
		"Content-Type": "application/json",
	}

	resp, err := req.Post(client.address+"/api/alert-notifications", header, postJSON)

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

	id, ok := responseBody["id"]

	if !ok {
		return "", errors.New("Response Body did not have id")
	}

	// this is really dumb, there has to be a better way to just get a string
	idFloat, ok := id.(float64)
	idInt := int(idFloat)

	if !ok {
		return "", fmt.Errorf("Unable to convert id %T %#v to int", id, id)
	}

	return strconv.Itoa(idInt), err
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

func responseIsSuccess(resp *req.Resp) bool {
	return resp.Response().StatusCode < 300 && resp.Response().StatusCode >= 200
}
