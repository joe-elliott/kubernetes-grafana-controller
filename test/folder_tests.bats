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
