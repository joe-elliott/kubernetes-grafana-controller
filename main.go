package main

import (
	"kubernetes-grafana-controller/pkg/signals"
	"os"
	"time"

	logging "github.com/op/go-logging"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	_log       = logging.MustGetLogger("prometheus-autoscaler")
	_logFormat = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{level:.4s} %{message}`,
	)
)

func init() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatted := logging.NewBackendFormatter(backend, _logFormat)

	logging.SetBackend(backendFormatted)
}

func main() {
	_log.Infof("Application Starting")

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		_log.Fatalf("rest.InClusterConfig failed: %v", err)
	}
	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		_log.Fatalf("kubernetes.NewForConfig failed: %v", err)
	}

	informerFactory := informers.NewSharedInformerFactory(client, time.Second*30)

	stopCh := signals.SetupSignalHandler()

	informerFactory.Start(stopCh)
}
