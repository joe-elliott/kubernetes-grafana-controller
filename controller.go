/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"kubernetes-grafana-controller/pkg/apis/samplecontroller/v1alpha1"
	samplev1alpha1 "kubernetes-grafana-controller/pkg/apis/samplecontroller/v1alpha1"
	clientset "kubernetes-grafana-controller/pkg/client/clientset/versioned"
	samplescheme "kubernetes-grafana-controller/pkg/client/clientset/versioned/scheme"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions/samplecontroller/v1alpha1"
	listers "kubernetes-grafana-controller/pkg/client/listers/samplecontroller/v1alpha1"
	"kubernetes-grafana-controller/pkg/grafana"
)

const controllerAgentName = "sample-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a GrafanaDashboard is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a GrafanaDashboard fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by GrafanaDashboard"
	// MessageResourceSynced is the message used for an Event fired when a GrafanaDashboard
	// is synced successfully
	MessageResourceSynced = "GrafanaDashboard synced successfully"
)

// Controller is the controller implementation for GrafanaDashboard resources
type Controller struct {
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	grafanaDashboardsLister listers.GrafanaDashboardLister
	grafanaDashboardsSynced cache.InformerSynced

	grafanaClient grafana.Interface

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	sampleclientset clientset.Interface,
	kubeclientset kubernetes.Interface,
	grafanaClient grafana.Interface,
	grafanaDashboardInformer informers.GrafanaDashboardInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		sampleclientset:         sampleclientset,
		grafanaDashboardsLister: grafanaDashboardInformer.Lister(),
		grafanaDashboardsSynced: grafanaDashboardInformer.Informer().HasSynced,
		grafanaClient:           grafanaClient,
		workqueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "GrafanaDashboards"),
		recorder:                recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when GrafanaDashboard resources change
	grafanaDashboardInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueGrafanaDashboard,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueGrafanaDashboard(new)
		},
		DeleteFunc: controller.enqueueGrafanaDashboard,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting GrafanaDashboard controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.grafanaDashboardsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process GrafanaDashboard resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var item WorkQueueItem
		var ok bool

		// confirm we have a work queue item
		if item, ok = obj.(WorkQueueItem); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// GrafanaDashboard resource to be synced.
		if err := c.syncHandler(item); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(item)
			return fmt.Errorf("error syncing '%s': %s, requeuing", item.key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until anothePostDashboardr change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", item.key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the GrafanaDashboard resource
// with the current status of the resource.
func (c *Controller) syncHandler(item WorkQueueItem) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(item.key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", item.key))
		return nil
	}

	// Get the GrafanaDashboard resource with this namespace/name
	grafanaDashboard, err := c.grafanaDashboardsLister.GrafanaDashboards(namespace).Get(name)
	if err != nil {
		// The GrafanaDashboard resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("grafanaDashboard '%s' in work queue no longer exists", item.key))
			return nil
		}

		return err
	}

	uid, err := c.grafanaClient.PostDashboard(grafanaDashboard.Spec.DashboardJSON)

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the GrafanaDashboard resource to reflect the
	// current state of the world
	err = c.updateGrafanaDashboardStatus(grafanaDashboard, uid)
	if err != nil {
		return err
	}

	c.recorder.Event(grafanaDashboard, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateGrafanaDashboardStatus(grafanaDashboard *samplev1alpha1.GrafanaDashboard, uid string) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	grafanaDashboardCopy := grafanaDashboard.DeepCopy()
	grafanaDashboardCopy.Status.GrafanaUID = uid
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the GrafanaDashboard resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.SamplecontrollerV1alpha1().GrafanaDashboards(grafanaDashboard.Namespace).Update(grafanaDashboardCopy)
	return err
}

// enqueueGrafanaDashboard takes a GrafanaDashboard resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than GrafanaDashboard.
func (c *Controller) enqueueGrafanaDashboard(obj interface{}) {
	var key string
	var err error
	var dashboard *v1alpha1.GrafanaDashboard
	var ok bool

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	if dashboard, ok = obj.(*v1alpha1.GrafanaDashboard); !ok {
		utilruntime.HandleError(fmt.Errorf("expected GrafanaDashboard in workqueue but got %#v", obj))
		return
	}

	item := NewWorkQueueItem(key, Dashboard, dashboard.Status.GrafanaUID) // todo: confirm this doesnt need null checking

	c.workqueue.AddRateLimited(item)
}
