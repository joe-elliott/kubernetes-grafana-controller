package controllers

/*
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

func TestCreatesGrafanaNotificationChannel(t *testing.T) {

	f := newFixture(t)
	channelJson := "{ 'test': 'test' }"

	channel := newGrafanaNotificationChannel("test", channelJson)

	f.grafanaChannelLister = append(f.grafanaChannelLister, channel)
	f.objects = append(f.objects, channel)

	channel.Status.GrafanaID = FAKE_UID
	f.expectUpdateGrafanaUid(dashboard)
	f.expectGrafanaDashboardPost(dashboardJson)

	f.run(getKey(channel, t))
}*/
