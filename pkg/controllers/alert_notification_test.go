package controllers

import (
	"github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/clientset/versioned/fake"
	informers "github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/informers/externalversions"
	grafana "github.com/joe-elliott/kubernetes-grafana-controller/pkg/grafana/fake"
	"testing"

	grafanacontroller "github.com/joe-elliott/kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

func newGrafanaAlertNotification(name string, notificationJson string) *grafanacontroller.GrafanaAlertNotification {
	return &grafanacontroller.GrafanaAlertNotification{
		TypeMeta: metav1.TypeMeta{APIVersion: grafanacontroller.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: grafanacontroller.GrafanaAlertNotificationSpec{
			AlertNotificationJSON: notificationJson,
		},
	}
}

func newAlertNotificationController(f *fixture) (*Controller, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.grafanaClient = grafana.NewGrafanaClientFake("https://example.com", FAKE_UID)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewAlertNotificationController(f.client, f.kubeclient,
		f.grafanaClient, i.Grafana().V1alpha1().GrafanaAlertNotifications())

	c.informerSynced = alwaysReady
	c.syncer.(*AlertNotificationSyncer).recorder = &record.FakeRecorder{}

	for _, d := range f.grafanaNotificationLister {
		i.Grafana().V1alpha1().GrafanaAlertNotifications().Informer().GetIndexer().Add(d)
	}

	return c, i
}

func TestCreatesGrafanaAlertNotification(t *testing.T) {

	f := newFixture(t)
	notificationJson := "{ 'test': 'test' }"

	notification := newGrafanaAlertNotification("test", notificationJson)
	item := NewWorkQueueItem(getKey(notification, t), nil, "")

	f.grafananotificationLister = append(f.grafananotificationLister, notification)
	f.objects = append(f.objects, notification)

	notification.Status.GrafanaID = FAKE_UID
	f.expectUpdateGrafanaObject(notification, notification.Namespace, "grafanaalertnotifications")
	f.expectGrafanaPost(notificationJson)

	f.runController(newAlertNotificationController, item, true, false)
}
