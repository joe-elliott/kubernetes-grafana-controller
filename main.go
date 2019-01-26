package main

import (
	"flag"
	"sync"
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
	masterURL          string
	kubeconfig         string
	grafanaURL         string
	resyncDeletePeriod time.Duration
	resyncPeriod       time.Duration
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&grafanaURL, "grafana", "http://grafana", "The address of the Grafana server.")
	flag.DurationVar(&resyncDeletePeriod, "resync-delete", time.Second*30, "Periodic interval in which to force resync deleted objects.  Pass 0s to disable.")
	flag.DurationVar(&resyncPeriod, "resync", time.Second*30, "Periodic interval in which to force resync objects.")

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
	informerFactory := informers.NewSharedInformerFactory(client, resyncPeriod)

	var allControllers []*controllers.Controller

	allControllers = append(allControllers, controllers.NewDashboardController(client,
		kubeClient,
		grafanaClient,
		informerFactory.Grafana().V1alpha1().Dashboards()))

	allControllers = append(allControllers, controllers.NewAlertNotificationController(client,
		kubeClient,
		grafanaClient,
		informerFactory.Grafana().V1alpha1().AlertNotifications()))

	allControllers = append(allControllers, controllers.NewDataSourceController(client,
		kubeClient,
		grafanaClient,
		informerFactory.Grafana().V1alpha1().DataSources()))

	stopCh := signals.SetupSignalHandler()

	informerFactory.Start(stopCh)

	var wg sync.WaitGroup

	for _, controller := range allControllers {
		wg.Add(1)

		go func(c *controllers.Controller) {
			defer wg.Done()

			if err := c.Run(2, resyncDeletePeriod, stopCh); err != nil {
				klog.Fatalf("Error running controller: %s", err.Error())
			}
		}(controller)
	}

	wg.Wait()
}
