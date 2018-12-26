package test

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	nosetup = flag.Bool("nosetup", false, "Skip minikube/grafana setup")
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
}

func teardownIntegration() {
	fmt.Println("teardownIntegration")
}

//
// tests
//

func TestDashboardPost(t *testing.T) {
	fmt.Println("TestDashboardPost")

	t.Error("Fail")
}
