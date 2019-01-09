package controllers

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"

	grafanacontroller "kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	"kubernetes-grafana-controller/pkg/client/clientset/versioned/fake"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions"
	grafana "kubernetes-grafana-controller/pkg/grafana/fake"
)

func newGrafanaDashboard(name string, dashboardJson string) *grafanacontroller.GrafanaDashboard {
	return &grafanacontroller.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{APIVersion: grafanacontroller.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: grafanacontroller.GrafanaDashboardSpec{
			DashboardJSON: dashboardJson,
		},
	}
}

func newDashboardController(f *fixture) (*Controller, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.grafanaClient = grafana.NewGrafanaClientFake("https://example.com", FAKE_UID)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewDashboardController(f.client, f.kubeclient,
		f.grafanaClient, i.Grafana().V1alpha1().GrafanaDashboards())

	c.informerSynced = alwaysReady
	c.syncer.(*DashboardSyncer).recorder = &record.FakeRecorder{}

	for _, d := range f.grafanaDashboardLister {
		i.Grafana().V1alpha1().GrafanaDashboards().Informer().GetIndexer().Add(d)
	}

	return c, i
}

func TestCreatesGrafanaDashboard(t *testing.T) {

	f := newFixture(t)
	dashboardJson := "{ 'test': 'test' }"

	dashboard := newGrafanaDashboard("test", dashboardJson)
	item := NewWorkQueueItem(getKey(dashboard, t), Dashboard, "")

	f.grafanaDashboardLister = append(f.grafanaDashboardLister, dashboard)
	f.objects = append(f.objects, dashboard)

	dashboard.Status.GrafanaUID = FAKE_UID
	f.expectUpdateGrafanaObject(dashboard, dashboard.Namespace, "grafanadashboards")
	f.expectGrafanaPost(dashboardJson)

	f.runController(newDashboardController, item, true, false)
}
