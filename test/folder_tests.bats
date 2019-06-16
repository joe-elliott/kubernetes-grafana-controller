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

#
# folders
#
@test "creating a Folder object creates a Grafana Folder" {
    count=0

    for filename in folders/*.yaml; do
        folderId=$(validatePostFolder $filename)

        echo "Test Creating $filename ($folderId)"

        (( count++ ))
        validateFolderCount $count

        validateEvents Folder Synced $(objectNameFromFile $filename)
        
        validateMetrics grafana_controller_grafana_post_latency_ms folder
        validateMetrics grafana_controller_updated_object_total folder
    done

    validateMetrics grafana_controller_error_total 0
}

@test "deleting a Folder object deletes the Grafana Folder" {
    for filename in folders/*.yaml; do
        folderId=$(validatePostFolder $filename)

        echo "Test Deleting $filename ($folderId)"

        kubectl delete -f $filename

        sleep 5s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/folders/${folderId})

        [ "$httpStatus" -eq "404" ]

        validateFolderCount 0

        validateEvents Folder Synced $(objectNameFromFile $filename)
        validateEvents Folder Deleted $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms folder
        validateMetrics grafana_controller_updated_object_total folder
        validateMetrics grafana_controller_grafana_delete_latency_ms folder
        validateMetrics grafana_controller_deleted_object_total folder
    done

    validateMetrics grafana_controller_error_total 0
}

@test "deleting a Folder while the controller is not running deletes the folder in Grafana" {
    for filename in folders/*.yaml; do
        folderId=$(validatePostFolder $filename)

        kubectl scale --replicas=0 deployment/kubernetes-grafana-test

        echo "Test Deleting $filename ($folderId)"

        kubectl delete -f $filename

        sleep 5s

        kubectl scale --replicas=1 deployment/kubernetes-grafana-test

        sleep 10s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/folders/${folderId})

        [ "$httpStatus" -eq "404" ]

        validateFolderCount 0

        validateEvents Folder Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_resynced_deleted_total folder
    done

    validateMetrics grafana_controller_error_total 0
}

@test "creating a Folder object creates the same folder in Grafana" {
    count=0

    for filename in folders/*.yaml; do
        validateFolderContents $filename

        (( count++ ))
        validateFolderCount $count

        validateEvents Folder Synced $(objectNameFromFile $filename)
    done

    validateMetrics grafana_controller_error_total 0
}

@test "updating a Folder object updates the folder in Grafana" {
    count=0
    
    for filename in folders/*.yaml; do
        validateFolderContents $filename

        (( count++ ))
        validateFolderCount $count

        validateEvents Folder Synced $(objectNameFromFile $filename)
    done

    # the .update files have folders with the same ids and different contents. 
    #  not the best.  not the worst.  could be improved.
    for filename in folders/*.update; do
        validateFolderContents $filename

        validateFolderCount $count

        validateEvents Folder Synced $(objectNameFromFile $filename)

        validateMetrics grafana_controller_grafana_post_latency_ms folder
        validateMetrics grafana_controller_updated_object_total folder
    done

    validateMetrics grafana_controller_error_total 0
}

@test "state is resynced after deleting a folder in grafana" {
    for filename in folders/*.yaml; do
        folderId=$(validatePostFolder $filename)

        httpStatus=$(curl -X DELETE --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/folders/${folderId})

        [ "$httpStatus" -eq "200" ]

        sleep 30s

        httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/folders/${folderId})

        [ "$httpStatus" -eq "200" ]
    done

    validateMetrics grafana_controller_error_total 0
}