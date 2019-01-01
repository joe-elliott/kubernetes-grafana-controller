#!/usr/bin/env bats

load bats_utils

setup(){
    run kubectl scale --replicas=1 deployment/kubernetes-grafana-test
    kubectl apply -f grafana.yaml

    validateGrafanaUrl
}

teardown(){
    run kubectl scale --replicas=0 deployment/kubernetes-grafana-test

    run kubectl delete --ignore-not-found=true -f grafana.yaml

    kubectl delete GrafanaDashboard --ignore-not-found=true --all
    kubectl delete GrafanaNotificationChannel --ignore-not-found=true --all
}

#
# dashboards
#
@test "creating a GrafanaDashboard object creates a Grafana Dashboard" {
    count=0

    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        echo "Test Creating $filename ($dashboardId)"

        (( count++ ))
        validateDashboardCount $count
    done
}

@test "deleting a GrafanaDashboard object deletes the Grafana Dashboard" {

    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        echo "Test Deleting $filename ($dashboardId)"

        kubectl delete -f $filename

        sleep 5s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
        [ "$httpStatus" -eq "404" ]

        validateDashboardCount 0
    done
}

@test "creating a GrafanaDashboard object creates the same dashboard in Grafana" {
    count=0

    for filename in dashboards/*.yaml; do
        validateDashboardContents $filename

        (( count++ ))
        validateDashboardCount $count
    done
}

@test "updating a GrafanaDashboard object updates the dashboard in Grafana" {
    count=0
    
    for filename in dashboards/*.yaml; do
        validateDashboardContents $filename

        (( count++ ))
        validateDashboardCount $count
    done

    # the .update files have dashboards with the same ids and different contents. 
    #  not the best.  not the worst.  could be improved.
    for filename in dashboards/*.update; do
        validateDashboardContents $filename

        validateDashboardCount $count
    done
}

#
# notification channels
#
@test "creating a GrafanaNotificationChannel object creates a Grafana Notification Channel" {
    count=0

    for filename in notification_channels/*.yaml; do
        channelId=$(validatePostNotificationChannel $filename)

        echo "Test Creating $filename ($channelId)"

        (( count++ ))
        validateNotificationChannelCount $count
    done
}

@test "deleting a GrafanaNotificationChannel object deletes the Grafana Notification Channel" {

    for filename in notification_channels/*.yaml; do
        channelId=$(validatePostDashboard $filename)

        echo "Test Deleting $filename ($channelId)"

        kubectl delete -f $filename

        sleep 5s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${channelId})
        [ "$httpStatus" -eq "404" ]

        validateNotificationChannelCount 0
    done
}

@test "creating a GrafanaNotificationChannel object creates the same channel in Grafana" {
    count=0

    for filename in notification_channels/*.yaml; do
        validateNotificationChannelContents $filename

        (( count++ ))
        validateNotificationChannelCount $count
    done
}

@test "updating a GrafanaNotificationChannel object updates the channel in Grafana" {
    count=0
    
    for filename in dashboards/*.yaml; do
        validateNotificationChannelContents $filename

        (( count++ ))
        validateNotificationChannelCount $count
    done

    for filename in dashboards/*.update; do
        validateNotificationChannelContents $filename

        validateNotificationChannelCount $count
    done
}