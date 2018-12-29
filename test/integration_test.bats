#!/usr/bin/env bats

load bats_utils

setup(){
    run kubectl delete po -l run=kubernetes-grafana-test
    kubectl apply -f grafana.yaml

    validateGrafanaUrl
}

teardown(){
    run kubectl delete --ignore-not-found=true -f grafana.yaml

    for filename in dashboards/*; do
        run kubectl delete --ignore-not-found=true -f $filename
    done
}

@test "creating a GrafanaDashboard CRD creates a Grafana Dashboard" {
    count=0

    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        echo "Test Creating $filename ($dashboardId)"

        (( count++ ))
        validateDashboardCount $count
    done
}

@test "deleting a GrafanaDashboard CRD deletes the Grafana Dashboard" {

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

@test "creating a GrafanaDashboard CRD creates the same dashboard in Grafana" {
    count=0

    for filename in dashboards/*.yaml; do
        validateDashboardContents $filename

        (( count++ ))
        validateDashboardCount $count
    done
}

@test "updating a GrafanaDashboard CRD updates the same dashboard in Grafana" {
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