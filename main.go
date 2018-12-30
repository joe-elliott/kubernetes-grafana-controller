package main

import (
	"flag"
	"time"

	clientset "kubernetes-grafana-controller/pkg/client/clientset/versioned"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions"
	"kubernetes-grafana-controller/pkg/controllers"
	"kubernetes-grafana-controller/pkg/grafana"
	"kubernetes-grafana-controller/pkg/signals"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	masterURL  string
	kubeconfig string
	grafanaURL string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&grafanaURL, "grafana", "http://grafana", "The address of the Grafana server.")

	klog.InitFlags(nil)
}

func main() {
	klog.Info("Application Starting")

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
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

	grafanaClient := grafana.NewClient(grafanaURL)

	informerFactory := informers.NewSharedInformerFactory(client, time.Second*30)

	dashboardController := controllers.NewDashboardController(client, kubeClient, grafanaClient,
		informerFactory.Grafana().V1alpha1().GrafanaDashboards())

	stopCh := signals.SetupSignalHandler()

	informerFactory.Start(stopCh)

	if err = dashboardController.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running dashboardController: %s", err.Error())
	}
}
