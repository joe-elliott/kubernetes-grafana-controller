package test

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/imroc/req"
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

	time.Sleep(5 * time.Second)

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

	// dashboard exists!
}

func TestDashboardDelete(t *testing.T) {
	fmt.Println("TestDashboardDelete")

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

	time.Sleep(5 * time.Second)

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

	// dashboard exists!  now delete
	if err = run("kubectl", []string{"delete", "-f", "sample-dashboards.yaml"}); err != nil {
		t.Error("Failed to delete dashboards", err)
		return
	}

	time.Sleep(5 * time.Second)

	fmt.Println("Getting dashboard at " + grafanaUrl + "/api/dashboards/uid/" + id)
	if resp, err = req.Get(grafanaUrl + "/api/dashboards/uid/" + id); err != nil {
		t.Error("Failed to get dashboard", err)
		return
	}

	status = resp.Response().StatusCode
	if status != 404 {
		t.Error("Delete Dashboard status is not 404", resp.Response().Status)
		return
	}
}
