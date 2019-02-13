#!/usr/bin/env bats

load bats_utils

setup(){

 #   if [ "$BATS_TEST_NUMBER" -eq "1" ]; then
 #       teardown
 #   fi

    run kubectl scale --replicas=1 deployment/kubernetes-grafana-test
    run kubectl scale --replicas=1 deployment/grafana

    validateGrafanaUrl
    validateControllerUrl
}

teardown(){
    dumpState

    kubectl delete events --all

    run kubectl scale --replicas=0 deployment/kubernetes-grafana-test
    run kubectl scale --replicas=0 deployment/grafana

    kubectl delete Dashboard --ignore-not-found=true --all
    kubectl delete AlertNotification --ignore-not-found=true --all
    kubectl delete DataSource --ignore-not-found=true --all

    # clean up comparison files if they exist
    rm -f a.json
    rm -f b.json
}

#
# dashboards
#
@test "creating a Dashboard object creates a Grafana Dashboard" {
    count=0

    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        echo "Test Creating $filename ($dashboardId)"

        (( count++ ))
        validateDashboardCount $count

        validateEvents Dashboard Synced $(objectNameFromFile $filename)
        
        validateMetrics grafana_controller_grafana_post_latency_ms dashboard
        validateMetrics grafana_controller_updated_object_total dashboard
    done
}

@test "deleting a Dashboard object deletes the Grafana Dashboard" {
    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        echo "Test Deleting $filename ($dashboardId)"

        kubectl delete -f $filename

        sleep 5s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})

        [ "$httpStatus" -eq "404" ]

        validateDashboardCount 0

        validateEvents Dashboard Synced $(objectNameFromFile $filename)
        validateEvents Dashboard Deleted $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms dashboard
        validateMetrics grafana_controller_updated_object_total dashboard
        validateMetrics grafana_controller_grafana_delete_latency_ms dashboard
        validateMetrics grafana_controller_deleted_object_total dashboard
    done
}

@test "deleting a Dashboard while the controller is not running deletes the dashboard in Grafana" {
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

        validateEvents Dashboard Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_resynced_deleted_total dashboard
    done
}

@test "creating a Dashboard object creates the same dashboard in Grafana" {
    count=0

    for filename in dashboards/*.yaml; do
        validateDashboardContents $filename

        (( count++ ))
        validateDashboardCount $count

        validateEvents Dashboard Synced $(objectNameFromFile $filename)
    done
}

@test "updating a Dashboard object updates the dashboard in Grafana" {
    count=0
    
    for filename in dashboards/*.yaml; do
        validateDashboardContents $filename

        (( count++ ))
        validateDashboardCount $count

        validateEvents Dashboard Synced $(objectNameFromFile $filename)
    done

    # the .update files have dashboards with the same ids and different contents. 
    #  not the best.  not the worst.  could be improved.
    for filename in dashboards/*.update; do
        validateDashboardContents $filename

        validateDashboardCount $count

        validateEvents Dashboard Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms dashboard
        validateMetrics grafana_controller_updated_object_total dashboard
    done
}

@test "state is resynced after deleting a dashboard in grafana" {
    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        httpStatus=$(curl -X DELETE --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})

        [ "$httpStatus" -eq "200" ]

        sleep 30s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})

        [ "$httpStatus" -eq "200" ]
    done
}

#
# alert notifications
#
@test "creating a AlertNotification object creates a Grafana Alert Notification" {
    for filename in alert_notifications/*.yaml; do
        notificationId=$(validatePostAlertNotification $filename)

        echo "Test Creating $filename ($notificationId)"

        (( count++ ))
        validateAlertNotificationCount $count

        validateEvents AlertNotification Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms alert-notification
        validateMetrics grafana_controller_updated_object_total alert-notification
    done
}

@test "deleting a AlertNotification object deletes the Grafana AlertNotification" {

    for filename in alert_notifications/*.yaml; do
        notificationId=$(validatePostAlertNotification $filename)

        echo "Test Deleting $filename ($notificationId)"

        kubectl delete -f $filename

        sleep 5s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${notificationId})

        # for some reason grafana 500s when you GET a non-existent alert-notifications?
        [ "$httpStatus" -eq "500" ]

        validateAlertNotificationCount 0

        validateEvents AlertNotification Synced $(objectNameFromFile $filename)
        validateEvents AlertNotification Deleted $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms alert-notification
        validateMetrics grafana_controller_updated_object_total alert-notification
        validateMetrics grafana_controller_grafana_delete_latency_ms alert-notification
        validateMetrics grafana_controller_deleted_object_total alert-notification
    done
}

@test "deleting a AlertNotification while the controller is not running deletes the alert notification in Grafana" {

    for filename in alert_notifications/*.yaml; do
        notificationId=$(validatePostAlertNotification $filename)

        kubectl scale --replicas=0 deployment/kubernetes-grafana-test

        echo "Test Deleting $filename ($notificationId)"

        kubectl delete -f $filename

        sleep 5s

        kubectl scale --replicas=1 deployment/kubernetes-grafana-test

        sleep 10s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${notificationId})

        # for some reason grafana 500s when you GET a non-existent alert-notifications?
        [ "$httpStatus" -eq "500" ]

        validateAlertNotificationCount 0

        validateEvents AlertNotification Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_resynced_deleted_total alert-notification
    done
}


@test "creating a AlertNotification object creates the same notification in Grafana" {
    count=0

    for filename in alert_notifications/*.yaml; do
        validateAlertNotificationContents $filename

        (( count++ ))
        validateAlertNotificationCount $count

        validateEvents AlertNotification Synced $(objectNameFromFile $filename)
    done
}

@test "updating a AlertNotification object updates the notification in Grafana" {
    count=0
    
    for filename in alert_notifications/*.yaml; do
        validateAlertNotificationContents $filename

        (( count++ ))
        validateAlertNotificationCount $count

        validateEvents AlertNotification Synced $(objectNameFromFile $filename)
    done

    for filename in alert_notifications/*.update; do
        validateAlertNotificationContents $filename

        validateAlertNotificationCount $count

        validateEvents AlertNotification Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms alert-notification
        validateMetrics grafana_controller_grafana_put_latency_ms alert-notification
        validateMetrics grafana_controller_updated_object_total alert-notification
    done
}

@test "state is resynced after deleting an alert notification in grafana" {
    for filename in alert_notifications/*.yaml; do
        notificationId=$(validatePostAlertNotification $filename)

        httpStatus=$(curl -X DELETE --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${notificationId})

        [ "$httpStatus" -eq "200" ]

        sleep 30s

        response=$(curl -s ${GRAFANA_URL}/api/alert-notifications)
        count=$(echo $response | jq "[.[] | select(.name == \"$(objectNameFromFile $filename)\")] | length")

        [ "$count" -eq "1" ]
    done
}

#
# data sources
#
@test "creating a DataSource object creates a Grafana DataSource" {
    count=0

    for filename in datasources/*.yaml; do
        sourceId=$(validatePostDataSource $filename)

        echo "Test Creating $filename ($sourceId)"

        (( count++ ))
        validateDataSourceCount $count

        validateEvents DataSource Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms datasource
        validateMetrics grafana_controller_updated_object_total datasource
    done
}

@test "deleting a DataSource object deletes the Grafana DataSource" {

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

        validateEvents DataSource Synced $(objectNameFromFile $filename)
        validateEvents DataSource Deleted $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms datasource
        validateMetrics grafana_controller_updated_object_total datasource
        validateMetrics grafana_controller_grafana_delete_latency_ms datasource
        validateMetrics grafana_controller_deleted_object_total datasource
    done
}

@test "deleting a DataSource while the controller is not running deletes the datasource in Grafana" {

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

        validateEvents DataSource Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_resynced_deleted_total datasoure
    done
}

@test "creating a DataSource object creates the same datasource in Grafana" {
    count=0

    for filename in datasources/*.yaml; do
        validateDataSourceContents $filename

        (( count++ ))
        validateDataSourceCount $count

        validateEvents DataSource Synced $(objectNameFromFile $filename)
    done
}

@test "updating a DataSource object updates the datasource in Grafana" {
    count=0
    
    for filename in datasources/*.yaml; do
        validateDataSourceContents $filename

        (( count++ ))
        validateDataSourceCount $count

        validateEvents DataSource Synced $(objectNameFromFile $filename)
    done

    for filename in datasources/*.update; do
        validateDataSourceContents $filename

        validateDataSourceCount $count

        validateEvents DataSource Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms datasource
        validateMetrics grafana_controller_grafana_put_latency_ms datasource
        validateMetrics grafana_controller_updated_object_total datasource
    done
}

@test "state is resynced after deleting a datasoure in grafana" {
    for filename in datasources/*.yaml; do
        sourceId=$(validatePostDataSource $filename)

        httpStatus=$(curl -X DELETE --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/datasources/${sourceId})

        [ "$httpStatus" -eq "200" ]

        sleep 30s

        response=$(curl -s ${GRAFANA_URL}/api/datasources)
        echo $response
        count=$(echo $response | jq "[.[] | select(.name == \"$(objectNameFromFile $filename)\")] | length")

        [ "$count" -eq "1" ]
    done
}