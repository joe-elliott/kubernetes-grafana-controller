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
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	samplecontroller "kubernetes-grafana-controller/pkg/apis/samplecontroller/v1alpha1"
	"kubernetes-grafana-controller/pkg/client/clientset/versioned/fake"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions"
	grafana "kubernetes-grafana-controller/pkg/grafana/fake"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

const (
	FAKE_UID = "fakeUID"
)

type fixture struct {
	t *testing.T

	client        *fake.Clientset
	kubeclient    *k8sfake.Clientset
	grafanaClient *grafana.GrafanaClientFake
	// Objects to put in the store.
	grafanaDashboardLister []*samplecontroller.GrafanaDashboard
	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newGrafanaDashboard(name string) *samplecontroller.GrafanaDashboard {
	return &samplecontroller.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{APIVersion: samplecontroller.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: samplecontroller.GrafanaDashboardSpec{
			DashboardJSON: "{}",
		},
	}
}

func (f *fixture) newController() (*Controller, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.grafanaClient = grafana.NewGrafanaClientFake("https://example.com", FAKE_UID)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewController(f.client, f.kubeclient,
		f.grafanaClient, i.Samplecontroller().V1alpha1().GrafanaDashboards())

	c.grafanaDashboardsSynced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	for _, d := range f.grafanaDashboardLister {
		i.Samplecontroller().V1alpha1().GrafanaDashboards().Informer().GetIndexer().Add(d)
	}

	return c, i
}

func (f *fixture) run(grafanaDashboardName string) {
	f.runController(grafanaDashboardName, true, false)
}

func (f *fixture) runExpectError(grafanaDashboardName string) {
	f.runController(grafanaDashboardName, true, true)
}

func (f *fixture) runController(grafanaDashboardName string, startInformers bool, expectError bool) {
	c, i := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
	}

	err := c.syncHandler(NewWorkQueueItem(grafanaDashboardName, Dashboard, ""))
	if !expectError && err != nil {
		f.t.Errorf("error syncing foo: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing foo, got nil")
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
				action.Matches("watch", "grafanadashboards")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateGrafanaUid(grafanaDashboard *samplecontroller.GrafanaDashboard) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "grafanadashboards"}, grafanaDashboard.Namespace, grafanaDashboard)
	// TODO: Until #38113 is merged, we can't use Subresource
	//action.Subresource = "status"
	f.actions = append(f.actions, action)
}

func getKey(grafanaDashboard *samplecontroller.GrafanaDashboard, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(grafanaDashboard)
	if err != nil {
		t.Errorf("Unexpected error getting key for foo %v: %v", grafanaDashboard.Name, err)
		return ""
	}
	return key
}

func TestCreatesGrafanaDashboard(t *testing.T) {
	// jpe - todo - add test for grafana client getting PostDashboard with the right json called
	f := newFixture(t)
	dashboard := newGrafanaDashboard("test")

	f.grafanaDashboardLister = append(f.grafanaDashboardLister, dashboard)
	f.objects = append(f.objects, dashboard)

	dashboard.Status.GrafanaUID = FAKE_UID
	f.expectUpdateGrafanaUid(dashboard)

	f.run(getKey(dashboard, t))
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	dashboard := newGrafanaDashboard("test")
	//d := newDeployment(foo)

	f.grafanaDashboardLister = append(f.grafanaDashboardLister, dashboard)
	f.objects = append(f.objects, dashboard)
	//f.deploymentLister = append(f.deploymentLister, d)
	//f.kubeobjects = append(f.kubeobjects, d)

	f.expectUpdateGrafanaUid(dashboard)
	f.run(getKey(dashboard, t))
}

func TestUpdateDashboard(t *testing.T) {
	f := newFixture(t)
	dashboard := newGrafanaDashboard("test")
	// d := newDeployment(foo)

	// Update dashboard
	dashboard.Spec.DashboardJSON = "{'test': 'test'}"
	//expDeployment := newDeployment(foo)

	f.grafanaDashboardLister = append(f.grafanaDashboardLister, dashboard)
	f.objects = append(f.objects, dashboard)
	//f.deploymentLister = append(f.deploymentLister, d)
	//f.kubeobjects = append(f.kubeobjects, d)

	f.expectUpdateGrafanaUid(dashboard)
	//f.expectUpdateDeploymentAction(expDeployment)
	f.run(getKey(dashboard, t))
}
