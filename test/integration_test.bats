#!/usr/bin/env bats

load bats_utils

setup(){

 #   if [ "$BATS_TEST_NUMBER" -eq "1" ]; then
 #       teardown
 #   fi

    run kubectl scale --replicas=1 deployment/kubernetes-grafana-test
    run kubectl scale --replicas=1 deployment/grafana

    validateGrafanaUrl
}

teardown(){
    dumpState

    kubectl delete events --all

    run kubectl scale --replicas=0 deployment/kubernetes-grafana-test
    run kubectl scale --replicas=0 deployment/grafana

    kubectl delete GrafanaDashboard --ignore-not-found=true --all
    kubectl delete GrafanaNotificationChannel --ignore-not-found=true --all
    kubectl delete GrafanaDataSource --ignore-not-found=true --all

    # clean up comparison files if they exist
    rm -f a.json
    rm -f b.json
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

        validateEventCount GrafanaDashboard Synced $(objectNameFromFile $filename) 1
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

        validateEventCount GrafanaDashboard Synced $(objectNameFromFile $filename) 1
        validateEventCount GrafanaDashboard Deleted $(objectNameFromFile $filename) 1
    done
}

@test "deleting a GrafanaDashboard while the controller is not running deletes the dashboard in Grafana" {
    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        kubectl scale --replicas=0 deployment/kubernetes-grafana-test

        echo "Test Deleting $filename ($dashboardId)"

        kubectl delete -f $filename

        sleep 5s

        kubectl scale --replicas=1 deployment/kubernetes-grafana-test

        sleep 10s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})

        [ "$httpStatus" -eq "404" ]

        validateDashboardCount 0

        validateEventCount GrafanaDashboard Synced $(objectNameFromFile $filename) 1
    done
}

@test "creating a GrafanaDashboard object creates the same dashboard in Grafana" {
    count=0

    for filename in dashboards/*.yaml; do
        validateDashboardContents $filename

        (( count++ ))
        validateDashboardCount $count

        validateEventCount GrafanaDashboard Synced $(objectNameFromFile $filename) 1
    done
}

@test "updating a GrafanaDashboard object updates the dashboard in Grafana" {
    count=0
    
    for filename in dashboards/*.yaml; do
        validateDashboardContents $filename

        (( count++ ))
        validateDashboardCount $count

        validateEventCount GrafanaDashboard Synced $(objectNameFromFile $filename) 1
    done

    # the .update files have dashboards with the same ids and different contents. 
    #  not the best.  not the worst.  could be improved.
    for filename in dashboards/*.update; do
        validateDashboardContents $filename

        validateDashboardCount $count

        validateEventCount GrafanaDashboard Synced $(objectNameFromFile $filename) 2
    done
}

#
# notification channels
#
@test "creating a GrafanaNotificationChannel object creates a Grafana Notification Channel" {
    for filename in notification_channels/*.yaml; do
        channelId=$(validatePostNotificationChannel $filename)

        echo "Test Creating $filename ($channelId)"

        (( count++ ))
        validateNotificationChannelCount $count

        validateEventCount GrafanaNotificationChannel Synced $(objectNameFromFile $filename) 1
    done
}

@test "deleting a GrafanaNotificationChannel object deletes the Grafana Notification Channel" {

    for filename in notification_channels/*.yaml; do
        channelId=$(validatePostNotificationChannel $filename)

        echo "Test Deleting $filename ($channelId)"

        kubectl delete -f $filename

        sleep 5s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${channelId})

        # for some reason grafana 500s when you GET a non-existent alert-notifications?
        [ "$httpStatus" -eq "500" ]

        validateNotificationChannelCount 0

        validateEventCount GrafanaNotificationChannel Synced $(objectNameFromFile $filename) 1
        validateEventCount GrafanaNotificationChannel Deleted $(objectNameFromFile $filename) 1
    done
}

@test "deleting a GrafanaNotificationChannel while the controller is not running deletes the notification channel in Grafana" {

    for filename in notification_channels/*.yaml; do
        channelId=$(validatePostNotificationChannel $filename)

        kubectl scale --replicas=0 deployment/kubernetes-grafana-test

        echo "Test Deleting $filename ($channelId)"

        kubectl delete -f $filename

        sleep 5s

        kubectl scale --replicas=1 deployment/kubernetes-grafana-test

        sleep 10s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${channelId})

        # for some reason grafana 500s when you GET a non-existent alert-notifications?
        [ "$httpStatus" -eq "500" ]

        validateNotificationChannelCount 0

        validateEventCount GrafanaNotificationChannel Synced $(objectNameFromFile $filename) 1
    done
}


@test "creating a GrafanaNotificationChannel object creates the same channel in Grafana" {
    count=0

    for filename in notification_channels/*.yaml; do
        validateNotificationChannelContents $filename

        (( count++ ))
        validateNotificationChannelCount $count

        validateEventCount GrafanaNotificationChannel Synced $(objectNameFromFile $filename) 1
    done
}

@test "updating a GrafanaNotificationChannel object updates the channel in Grafana" {
    count=0
    
    for filename in notification_channels/*.yaml; do
        validateNotificationChannelContents $filename

        (( count++ ))
        validateNotificationChannelCount $count

        validateEventCount GrafanaNotificationChannel Synced $(objectNameFromFile $filename) 1
    done

    for filename in notification_channels/*.update; do
        validateNotificationChannelContents $filename

        validateNotificationChannelCount $count

        validateEventCount GrafanaNotificationChannel Synced $(objectNameFromFile $filename) 2
    done
}

#
# data sources
#
@test "creating a GrafanaDataSource object creates a Grafana DataSource" {
    count=0

    for filename in datasources/*.yaml; do
        sourceId=$(validatePostDataSource $filename)

        echo "Test Creating $filename ($sourceId)"

        (( count++ ))
        validateDataSourceCount $count

        validateEventCount GrafanaDataSource Synced $(objectNameFromFile $filename) 1
    done
}

@test "deleting a GrafanaDataSource object deletes the Grafana DataSource" {

    for filename in datasources/*.yaml; do
        sourceId=$(validatePostDataSource $filename)

        echo "Test Deleting $filename ($sourceId)"

        kubectl delete -f $filename

        sleep 5s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/datasources/${sourceId})

        echo "status $httpStatus"
        curl ${GRAFANA_URL}/api/datasources

        [ "$httpStatus" -eq "404" ]

        validateDataSourceCount 0

        validateEventCount GrafanaDataSource Synced $(objectNameFromFile $filename) 1
        validateEventCount GrafanaDataSource Deleted $(objectNameFromFile $filename) 1
    done
}

@test "deleting a GrafanaDataSource while the controller is not running deletes the datasource in Grafana" {

    for filename in datasources/*.yaml; do
        sourceId=$(validatePostDataSource $filename)

        kubectl scale --replicas=0 deployment/kubernetes-grafana-test

        echo "Test Deleting $filename ($sourceId)"

        kubectl delete -f $filename

        sleep 5s

        kubectl scale --replicas=1 deployment/kubernetes-grafana-test

        sleep 10s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/datasources/${sourceId})

        echo "status $httpStatus"
        curl ${GRAFANA_URL}/api/datasources

        [ "$httpStatus" -eq "404" ]

        validateDataSourceCount 0

        validateEventCount GrafanaDataSource Synced $(objectNameFromFile $filename) 1
    done
}

@test "creating a GrafanaDataSource object creates the same datasource in Grafana" {
    count=0

    for filename in datasources/*.yaml; do
        validateDataSourceContents $filename

        (( count++ ))
        validateDataSourceCount $count

        validateEventCount GrafanaDataSource Synced $(objectNameFromFile $filename) 1
    done
}

@test "updating a GrafanaDataSource object updates the datasource in Grafana" {
    count=0
    
    for filename in datasources/*.yaml; do
        validateDataSourceContents $filename

        (( count++ ))
        validateDataSourceCount $count

        validateEventCount GrafanaDataSource Synced $(objectNameFromFile $filename) 1
    done

    for filename in datasources/*.update; do
        validateDataSourceContents $filename

        validateDataSourceCount $count

        validateEventCount GrafanaDataSource Synced $(objectNameFromFile $filename) 2
    done
}