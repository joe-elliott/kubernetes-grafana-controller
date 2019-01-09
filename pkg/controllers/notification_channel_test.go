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

func newGrafanaNotificationChannel(name string, channelJson string) *grafanacontroller.GrafanaNotificationChannel {
	return &grafanacontroller.GrafanaNotificationChannel{
		TypeMeta: metav1.TypeMeta{APIVersion: grafanacontroller.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: grafanacontroller.GrafanaNotificationChannelSpec{
			NotificationChannelJSON: channelJson,
		},
	}
}

func newNotificationChannelController(f *fixture) (*Controller, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.grafanaClient = grafana.NewGrafanaClientFake("https://example.com", FAKE_UID)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewNotificationChannelController(f.client, f.kubeclient,
		f.grafanaClient, i.Grafana().V1alpha1().GrafanaNotificationChannels())

	c.informerSynced = alwaysReady
	c.syncer.(*NotificationChannelSyncer).recorder = &record.FakeRecorder{}

	for _, d := range f.grafanaChannelLister {
		i.Grafana().V1alpha1().GrafanaNotificationChannels().Informer().GetIndexer().Add(d)
	}

	return c, i
}

func TestCreatesGrafanaNotificationChannel(t *testing.T) {

	f := newFixture(t)
	channelJson := "{ 'test': 'test' }"

	channel := newGrafanaNotificationChannel("test", channelJson)
	item := NewWorkQueueItem(getKey(channel, t), nil, "")

	f.grafanaChannelLister = append(f.grafanaChannelLister, channel)
	f.objects = append(f.objects, channel)

	channel.Status.GrafanaID = FAKE_UID
	f.expectUpdateGrafanaObject(channel, channel.Namespace, "grafananotificationchannels")
	f.expectGrafanaPost(channelJson)

	f.runController(newNotificationChannelController, item, true, false)
}
