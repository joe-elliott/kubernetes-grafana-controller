#!/usr/bin/env bats

load bats_utils

setup(){
    run kubectl delete po -l run=kubernetes-grafana-test
    kubectl apply -f grafana.yaml

    validateGrafanaUrl
}

teardown(){
    run kubectl delete --ignore-not-found=true -f grafana.yaml

    for filename in dashboards/*.yaml; do
        run kubectl delete --ignore-not-found=true -f $filename
    done
}

@test "creating a GrafanaDashboard CRD creates a Grafana Dashboard" {
    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        echo "Test Creating $filename ($dashboardId)"
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
    done
}

@test "creating a GrafanaDashboard CRD creates the same dashboard in Grafana" {

    for filename in dashboards/*.yaml; do
        dashboardId=$(validatePostDashboard $filename)

        echo "Test Json Content of $filename ($dashboardId)"

        dashboardJsonFromGrafana=$(curl --silent ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})

        echo $dashboardJsonFromGrafana | jq '.dashboard | del(.version) | del(.id) | del (.uid)' > a.json

        dashboardJsonFromYaml=$(grep -A9999 'dashboardJson' $filename)
        dashboardJsonFromYaml=${dashboardJsonFromYaml%?}   # strip final quote
        dashboardJsonFromYaml=${dashboardJsonFromYaml#*\'} # strip up to and including the first quote

        echo $dashboardJsonFromYaml | jq 'del(.version) | del(.id) | del (.uid)' > b.json

        equal=$(jq --argfile a a.json --argfile b b.json -n '$a == $b')

        if [ "$equal" != "true" ]; then
            run diff <(jq -S . a.json) <(jq -S . b.json)
            echo $output
        fi

        [ "$equal" = "true" ]

        rm a.json
        rm b.json
    done
}