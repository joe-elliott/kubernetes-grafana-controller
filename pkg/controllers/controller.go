package controllers

import (
	"errors"
	"fmt"
	"kubernetes-grafana-controller/pkg/client/clientset/versioned/scheme"
	"kubernetes-grafana-controller/pkg/grafana"
	"kubernetes-grafana-controller/pkg/prometheus"
	"time"

	grafanascheme "kubernetes-grafana-controller/pkg/client/clientset/versioned/scheme"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	SuccessSynced         = "Synced"
	MessageResourceSynced = "Grafana Object synced successfully"

	SuccessDeleted         = "Deleted"
	MessageResourceDeleted = "Grafana Object deleted successfully"
)

type Controller struct {
	syncer         Syncer
	informerSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

func NewController(informer cache.SharedIndexInformer,
	kubeclientset kubernetes.Interface,
	syncer Syncer) *Controller {

	controllerAgentName := "grafana-" + syncer.getType() + "-controller"

	// Create event broadcaster
	// Add grafana-controller types to the default Kubernetes Scheme so Events can be
	// logged for grafana-controller types.
	utilruntime.Must(grafanascheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		workqueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerAgentName),
		informerSynced: informer.HasSynced,
		syncer:         syncer,
		recorder:       recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when GrafanaDashboard resources change
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(toAdd interface{}) {
			controller.enqueueWorkQueueItem(toAdd, AddOrUpdate)
		},
		UpdateFunc: func(old, new interface{}) {

			controller.enqueueWorkQueueItem(new, AddOrUpdate)
		},
		DeleteFunc: func(toDelete interface{}) {
			controller.enqueueWorkQueueItem(toDelete, Delete)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, resyncDeletePeriod time.Duration, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting GrafanaDashboard controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.informerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process GrafanaDashboard resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	// launch resync all thing
	if resyncDeletePeriod != 0 {
		go wait.Until(c.enqueueResyncDeletedObjects, resyncDeletePeriod, stopCh)
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

		if item.isResyncDeletedObjects() {
			prometheus.ResyncDeletedTotal.WithLabelValues(c.syncer.getType()).Inc()

			if err := c.resyncDeletedObjects(); err != nil {
				prometheus.ErrorTotal.Inc()

				c.workqueue.AddRateLimited(item)
				return fmt.Errorf("error resyncing all %s, requeuing", err.Error())
			}
		} else {

			if err := c.syncHandler(item); err != nil {
				prometheus.ErrorTotal.Inc()

				c.workqueue.AddRateLimited(item)
				return fmt.Errorf("error syncing '%s': %s, requeuing", item.key, err.Error())
			}
		}

		// remove from queue
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

func (c *Controller) syncHandler(item WorkQueueItem) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(item.key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", item.key))
		return nil
	}

	if item.itemType == Delete {
		// object was deleted, so delete from grafana
		err = c.syncer.deleteObjectById(item.id)

		if err == nil {
			prometheus.DeletedObjectTotal.WithLabelValues(c.syncer.getType()).Inc()
			c.recorder.Event(item.originalObject, corev1.EventTypeNormal, SuccessDeleted, MessageResourceDeleted)
		}

		return err
	}

	// Get the DataSource resource with this namespace/name
	runtimeObject, err := c.syncer.getRuntimeObjectByName(name, namespace)
	if err != nil {

		if k8serrors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("Grafana Object '%s' in work queue no longer exists? Deleting", item.key))
			prometheus.ErrorTotal.Inc()

			// object was deleted, so delete from grafana
			err = c.syncer.deleteObjectById(item.id)

			if err == nil {
				prometheus.DeletedObjectTotal.WithLabelValues(c.syncer.getType()).Inc()
				c.recorder.Event(item.originalObject, corev1.EventTypeNormal, SuccessDeleted, MessageResourceDeleted)
			}
		}

		return err
	}

	err = c.syncer.updateObject(runtimeObject)

	if err != nil {
		return err
	}

	prometheus.UpdatedObjectTotal.WithLabelValues(c.syncer.getType()).Inc()
	c.recorder.Event(runtimeObject, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) resyncDeletedObjects() error {

	// get all dashboards in grafana.  anything in grafana that's not in k8s gets nuked
	kubernetesIDs, err := c.syncer.getAllKubernetesObjectIDs()

	if err != nil {
		return err
	}

	grafanaIDs, err := c.syncer.getAllGrafanaObjectIDs()

	if err != nil {
		return err
	}

	for _, grafanaID := range grafanaIDs {
		var found = false

		for _, kubernetesID := range kubernetesIDs {

			if kubernetesID == grafana.NO_ID {
				return errors.New("found kubernetes object with unitialized id, bailing")
			}

			if kubernetesID == grafanaID {
				found = true
				break
			}
		}

		if !found {
			klog.Infof("Object found in grafana but not k8s.  Deleting")
			err = c.syncer.deleteObjectById(grafanaID)

			// if one fails just go ahead and bail out.  controlling logic will requeue
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Controller) enqueueResyncDeletedObjects() {
	c.workqueue.AddRateLimited(NewResyncDeletedObjects())
}

func (c *Controller) enqueueWorkQueueItem(obj interface{}, itemType WorkQueueItemType) {

	item := c.syncer.createWorkQueueItem(obj)
	item.itemType = itemType

	if item != nil {
		c.workqueue.AddRateLimited(*item)
	}
}
