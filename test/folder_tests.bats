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
    kubectl delete Folder --ignore-not-found=true --all

    # clean up comparison files if they exist
    rm -f a.json
    rm -f b.json
}

@test "creating a Dashboard in a Folder object creates a Grafana Dashboard in a Folder" {

    folderId=$(validatePostFolder 'folders/test.yaml')
    
    sleep 5s

    kubectl create -f folders/dash-folder.test

    sleep 5s

    echo "Test Creating folder ($folderId)"

    folderIntId=$(curl --silent ${GRAFANA_URL}/api/folders/${folderId} | jq '.id')

    echo "Checking folder int id $folderIntId"

    dashboardCount=$(curl --silent ${GRAFANA_URL}/api/search?folderIds=${folderIntId} | jq '. | length ')

    echo "dashboardCount: $dashboardCount"

    [ "$dashboardCount" = "1" ]
}