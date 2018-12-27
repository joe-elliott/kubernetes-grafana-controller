#!/usr/bin/env bats

load setup

setup(){
    # one time setup
    if [ "$BATS_TEST_NUMBER" -eq 1 ]; then
        setupIntegrationTests
    fi

    kubectl apply -f grafana.yaml

    validateGrafanaUrl

    # confirm one time setup worked before every step
    [ "$SETUP_SUCCEEDED" -eq "1" ]
}

teardown(){
    run kubectl delete --ignore-not-found=true -f sample-dashboards.yaml
    run kubectl delete --ignore-not-found=true -f grafana.yaml

    # one time teardown
    if [ "$BATS_TEST_NUMBER" -eq ${#BATS_TEST_NAMES[@]} ]; then
        teardownIntegrationTests
    fi
}

@test "Create Dashboard" {

    # create in kubernetes
    kubectl apply -f sample-dashboards.yaml

	sleep 5s

    getGrafanaDashboardIdByName test-dash
    dashboardId=$?

    echo "Grafana Dashboard Id " $dashboardId

    # check if exists in grafana
	$httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "200" ]
}

@test "Delete Dashboard" {

    # create in kubernetes
    kubectl apply -f sample-dashboards.yaml

	sleep 5s

    getGrafanaDashboardIdByName test-dash
    dashboardId=$?

    echo "Grafana Dashboard Id " $dashboardId

    # check if exists in grafana
	$httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "200" ]

    kubectl delete -f sample-dashboards.yaml

	sleep 5s

	$httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "404" ]
}