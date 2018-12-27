#!/usr/bin/env bats

load setup

setup(){
    # one time setup
    if [ "$BATS_TEST_NUMBER" -eq 1 ]; then
        setupIntegrationTests
    fi

    kubectl apply -f crd.yaml
    [ "$?" -eq "0" ]

    kubectl apply -f grafana.yaml
    [ "$?" -eq "0" ]

    validateGrafanaUrl

    # confirm one time setup worked before every step
    [ "$SETUP_SUCCEEDED" -eq "1" ]
}

teardown(){
    # one time teardown
    if [ "$BATS_TEST_NUMBER" -eq ${#BATS_TEST_NAMES[@]} ]; then
        teardownIntegrationTests
    fi

    kubectl delete -f crd.yaml
    [ "$?" -eq "0" ]

    kubectl delete -f grafana.yaml
    [ "$?" -eq "0" ]
}

@test "Create Dashboard" {

}