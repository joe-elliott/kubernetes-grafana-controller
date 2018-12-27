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
    # one time teardown
    if [ "$BATS_TEST_NUMBER" -eq ${#BATS_TEST_NAMES[@]} ]; then
        teardownIntegrationTests
    fi

    kubectl delete --ignore-not-found=true -f sample-dashboards.yaml
    kubectl delete --ignore-not-found=true -f grafana.yaml
}

@test "Create Dashboard" {

    # create in kubernetes
    kubectl apply -f sample-dashboards.yaml

	sleep 5s

    getGrafanaDashboardIdByName test-dash
    dashboardId=$?

    echo "Grafana Dashboard Id " $dashboardId

    # check if exists in grafana
	run curl --silent --output /dev/null --write-out "%{http_code}" $GRAFANA_URL
    [ "$status" -eq "200" ]
}