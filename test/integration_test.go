package test

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var (
	nosetup = flag.Bool("nosetup", false, "Skip minikube/grafana setup")
)

const (
	MINIKUBE_NAME = "minikube-grafana-test"
)

func TestMain(m *testing.M) {
	flag.Parse()

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

	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Failed:  Needs to be run as root to setup minikube")
		os.Exit(1)
	}

	// ignore failure on these.  they will fail if a minikube cluster does not exist
	run("minikube", []string{"stop", "-p", MINIKUBE_NAME})
	run("minikube", []string{"delete", "-p", MINIKUBE_NAME})

	if err := run("minikube", []string{"start", "-p", MINIKUBE_NAME}); err != nil {
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
}

func teardownIntegration() {
	fmt.Println("teardownIntegration")

	run("minikube", []string{"stop", "-p", MINIKUBE_NAME})
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

//
// tests
//

func TestDashboardPost(t *testing.T) {
	fmt.Println("TestDashboardPost")

	t.Error("Fail")
}
