#!/usr/bin/env bats

load bats_utils

setup(){
    run kubectl delete po -l run=kubernetes-grafana-test
    kubectl apply -f grafana.yaml

    validateGrafanaUrl
}

teardown(){
    run kubectl delete --ignore-not-found=true -f grafana.yaml
    run kubectl delete --ignore-not-found=true -f dashboards/test-dash.yaml
}

@test "creating a GrafanaDashboard CRD creates a Grafana Dashboard" {
    for filename in dashboards/*.yaml; do
        validatePostDashboard $filename
    done
}

@test "deleting a GrafanaDashboard CRD deletes the Grafana Dashboard" {

    # create in kubernetes
    kubectl apply -f dashboards/test-dash.yaml

	sleep 5s

    dashboardName="test-dash"
    dashboardId=$(kubectl get GrafanaDashboard -o=jsonpath="{.items[?(@.metadata.name==\"${dashboardName}\")].status.grafanaUID}")

    echo "Grafana Dashboard Id " $dashboardId

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "200" ]

    kubectl delete -f dashboards/test-dash.yaml

	sleep 5s

	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "404" ]
}

@test "creating a GrafanaDashboard CRD creates the same dashboard in Grafana" {

    # create in kubernetes
    kubectl apply -f dashboards/test-dash.yaml

	sleep 5s

    dashboardName="test-dash"
    dashboardId=$(kubectl get GrafanaDashboard -o=jsonpath="{.items[?(@.metadata.name==\"${dashboardName}\")].status.grafanaUID}")

    echo "Grafana Dashboard Id " $dashboardId

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "200" ]

    dashboardJsonFromGrafana=$(curl --silent ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})

    echo $dashboardJsonFromGrafana | jq '.dashboard | del(.version) | del(.id)' > a.json

    dashboardJsonFromYaml=$(grep -A9999 'dashboardJson' dashboards/test-dash.yaml)
    dashboardJsonFromYaml=${dashboardJsonFromYaml%?}   # strip final quote
    dashboardJsonFromYaml=${dashboardJsonFromYaml#*\'} # strip up to and including the first quote

    echo $dashboardJsonFromYaml | jq 'del(.version) | del(.id)' > b.json

    equal=$(jq --argfile a a.json --argfile b b.json -n '$a == $b')

    if [ "$equal" != "true" ]; then
        run diff <(jq -S . a.json) <(jq -S . b.json)
        echo $output
    fi

    [ "$equal" = "true" ]

    rm a.json
    rm b.json
}