package main

import (
	"flag"
	clientset "kubernetes-grafana-controller/pkg/client/clientset/versioned"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions"
	"kubernetes-grafana-controller/pkg/signals"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
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

	informerFactory := informers.NewSharedInformerFactory(client, time.Second*30)

	controller := NewController(client, kubeClient,
		informerFactory.Samplecontroller().V1alpha1().Foos())

	stopCh := signals.SetupSignalHandler()

	informerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
