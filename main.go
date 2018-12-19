package main

import (
	clientset "kubernetes-grafana-controller/pkg/client/clientset/versioned"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions"
	"kubernetes-grafana-controller/pkg/signals"
	"os"
	"time"

	logging "github.com/op/go-logging"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
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
	klog.Info("Application Starting")

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("rest.InClusterConfig failed: %v", err)
	}
	// creates the clientset
	client, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("clientset.NewForConfig failed: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(client, time.Second*30)

	controller := NewController(client, kubeClient,
		informerFactory.Samplecontroller().V1alpha1().Foos())

	stopCh := signals.SetupSignalHandler()

	informerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
