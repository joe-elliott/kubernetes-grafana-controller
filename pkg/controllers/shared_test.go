package controllers

import (
	grafanacontroller "kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	"kubernetes-grafana-controller/pkg/client/clientset/versioned/fake"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions"
	grafana "kubernetes-grafana-controller/pkg/grafana/fake"
	"reflect"
	"testing"
	"time"

	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	core "k8s.io/client-go/testing"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

const (
	FAKE_UID = "fakeUID"
)

type controllerFactory func(*fixture) (*Controller, informers.SharedInformerFactory)

type fixture struct {
	t *testing.T

	client        *fake.Clientset
	kubeclient    *k8sfake.Clientset
	grafanaClient *grafana.ClientFake

	// Objects to put in the store.
	grafanaDashboardLister  []*grafanacontroller.GrafanaDashboard
	grafanaChannelLister    []*grafanacontroller.GrafanaNotificationChannel
	grafanaDataSourceLister []*grafanacontroller.GrafanaDataSource

	// Actions expected to happen on the client.
	kubeactions       []core.Action
	actions           []core.Action
	grafanaPostedJson *string

	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	f.grafanaPostedJson = nil

	return f
}

func (f *fixture) runController(newController controllerFactory, item WorkQueueItem, startInformers bool, expectError bool) {
	c, i := newController(f)
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
	}

	err := c.syncer.syncHandler(item)
	if !expectError && err != nil {
		f.t.Errorf("error syncing dashboard: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing dashboard, got nil")
	}

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}

	k8sActions := filterInformerActions(f.kubeclient.Actions())
	for i, action := range k8sActions {
		if len(f.kubeactions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
			break
		}

		expectedAction := f.kubeactions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}

	// test grafana client "actions"
	if f.grafanaPostedJson != nil &&
		*f.grafanaPostedJson != *f.grafanaClient.PostedJson {

		f.t.Errorf("Expected grafana posted json %s but found %s", *f.grafanaPostedJson, *f.grafanaClient.PostedJson)
	}
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateAction:
		e, _ := expected.(core.CreateAction)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintDiff(expObject, object))
		}
	case core.UpdateAction:
		e, _ := expected.(core.UpdateAction)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintDiff(expObject, object))
		}
	case core.PatchAction:
		e, _ := expected.(core.PatchAction)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintDiff(expPatch, patch))
		}
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// noise level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {

	ret := []core.Action{}
	for _, action := range actions {

		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "grafanadashboards") ||
				action.Matches("watch", "grafanadashboards") ||
				action.Matches("list", "grafananotificationchannels") ||
				action.Matches("watch", "grafananotificationchannels")) {
			continue
		}
		ret = append(ret, action)
	}

	return actions
}

func (f *fixture) expectUpdateGrafanaObject(obj runtime.Object, namespace string, resource string) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: resource}, namespace, obj)
	// TODO: Until #38113 is merged, we can't use Subresource
	//action.Subresource = "status"
	f.actions = append(f.actions, action)
}

func (f *fixture) expectGrafanaPost(json string) {
	f.grafanaPostedJson = &json
}

func getKey(obj interface{}, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		t.Errorf("Unexpected error getting key for object %v: %v", obj, err)
		return ""
	}
	return key
}
