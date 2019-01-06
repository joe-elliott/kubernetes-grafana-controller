package controllers

import (
	"kubernetes-grafana-controller/pkg/client/clientset/versioned/fake"
	informers "kubernetes-grafana-controller/pkg/client/informers/externalversions"
	grafana "kubernetes-grafana-controller/pkg/grafana/fake"
	"testing"

	grafanacontroller "kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

func newGrafanaDataSource(name string, dataSourceJson string) *grafanacontroller.GrafanaDataSource {
	return &grafanacontroller.GrafanaDataSource{
		TypeMeta: metav1.TypeMeta{APIVersion: grafanacontroller.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: grafanacontroller.GrafanaDataSourceSpec{
			DataSourceJSON: dataSourceJson,
		},
	}
}

func newDataSourceController(f *fixture) (*Controller, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.grafanaClient = grafana.NewGrafanaClientFake("https://example.com", FAKE_UID)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewDataSourceController(f.client, f.kubeclient,
		f.grafanaClient, i.Grafana().V1alpha1().GrafanaDataSources())

	c.informerSynced = alwaysReady
	c.syncer.(*DataSourceSyncer).recorder = &record.FakeRecorder{}

	for _, d := range f.grafanaDataSourceLister {
		i.Grafana().V1alpha1().GrafanaDataSources().Informer().GetIndexer().Add(d)
	}

	return c, i
}

func TestCreatesGrafanaDataSource(t *testing.T) {

	f := newFixture(t)
	dataSourceJson := "{ 'test': 'test' }"

	dataSource := newGrafanaDataSource("test", dataSourceJson)
	item := NewWorkQueueItem(getKey(dataSource, t), DataSource, "")

	f.grafanaDataSourceLister = append(f.grafanaDataSourceLister, dataSource)
	f.objects = append(f.objects, dataSource)

	dataSource.Status.GrafanaID = FAKE_UID
	f.expectUpdateGrafanaObject(dataSource, dataSource.Namespace, "grafanadatasources")
	f.expectGrafanaPost(dataSourceJson)

	f.runController(newDataSourceController, item, true, false)
}
