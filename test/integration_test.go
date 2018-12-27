package test

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/imroc/req"
	yaml "gopkg.in/yaml.v2"
)

var (
	nosetup = flag.Bool("nosetup", false, "Skip minikube/grafana setup")
)

func TestMain(m *testing.M) {
	flag.Parse()

	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Failed:  Needs to be run as root to setup minikube")
		os.Exit(1)
	}

	if !(*nosetup) {
		setupIntegration()
	}

	result := m.Run()

	if !(*nosetup) {
		teardownIntegration()
	}
	os.Exit(result)
}

//
// integration test
//  - minikube
//  - crds
//  - grafana
//
func setupIntegration() {
	fmt.Println("setupIntegration")

	// ignore failure on these.  they will fail if a minikube cluster does not exist
	run("minikube", []string{"stop"})
	run("minikube", []string{"delete"})

	if err := run("minikube", []string{"start"}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := run("kubectl", []string{"create", "-f", "crd.yaml"}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := run("kubectl", []string{"create", "-f", "grafana.yaml"}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// build dockerfile and deploy to minikube
	if err := run("./integration_test.sh", []string{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func teardownIntegration() {
	fmt.Println("teardownIntegration")

	run("minikube", []string{"stop", "-p"})
}

func run(cmdName string, cmdArgs []string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func getGrafanaUrl() (string, error) {

	out, err := exec.Command("minikube", []string{"service", "grafana", "--url"}...).Output()
	if err != nil {
		return "", err
	}

	grafanaUrl := strings.TrimSpace(string(out))

	_, err = url.ParseRequestURI(grafanaUrl)

	if err != nil {
		return "", err
	}

	return grafanaUrl, nil
}

func getGrafanaDashboardId(name string) (string, error) {

	jsonPathArg := fmt.Sprintf("-o=jsonpath='{.items[?(@.metadata.name==\"%s\")].status.grafanaUID}'", name)

	out, err := exec.Command("kubectl", []string{"get", "GrafanaDashboard", jsonPathArg}...).Output()
	if err != nil {
		return "", err
	}

	id := strings.TrimSpace(string(out))
	id = strings.Trim(id, "'")

	if len(id) == 0 {
		return "", fmt.Errorf("Grafana Id is empty for %s", name)
	}

	return id, nil
}

func areEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

//
// tests
//

func TestDashboardPost(t *testing.T) {
	fmt.Println("TestDashboardPost")

	var id string
	var resp *req.Resp

	grafanaUrl, err := getGrafanaUrl()
	if err != nil {
		t.Error("Failed to get Grafana URL", err)
	}

	// create dashboard
	if err = run("kubectl", []string{"apply", "-f", "sample-dashboards.yaml"}); err != nil {
		t.Error("Failed to create dashboards", err)
		return
	}

	// get object status to get grafana id
	if id, err = getGrafanaDashboardId("test-dash"); err != nil {
		t.Error("Failed to get id", err)
		return
	}

	// GET grafana dashboard with id
	fmt.Println("Getting dashboard at " + grafanaUrl + "/api/dashboards/uid/" + id)
	if resp, err = req.Get(grafanaUrl + "/api/dashboards/uid/" + id); err != nil {
		t.Error("Failed to get dashboard", err)
		return
	}

	status := resp.Response().StatusCode
	if status >= 300 || status < 200 {
		t.Error("Get Dashboard status is unsuccessful", resp.Response().Status)
		return
	}

	// get file json
	var fileBytes []byte
	var grafanaDashboard map[interface{}]interface{}

	fileBytes, err = ioutil.ReadFile("sample-dashboards.yaml")
	if err != nil {
		t.Error("Error reading sample-dashboards.yaml", err)
		return
	}

	err = yaml.Unmarshal(fileBytes, &grafanaDashboard)
	if err != nil {
		t.Error("Error unmarshalling grafanaDashboard", err)
		return
	}

	var equal bool
	spec := grafanaDashboard["spec"]
	dashboardJson := spec.(map[interface{}]interface{})["dashboardJson"]
	equal, err = areEqualJSON(dashboardJson.(string), resp.String())

	if err != nil {
		t.Error("Error comparing json", err)
		return
	}

	if !equal {
		t.Error("Dashboard jsons are not equal")
	}
}
