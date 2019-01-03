export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false

GRAFANA_URL=""

setupIntegrationTests() {

    # ignore failure on these.  they will fail if a minikube cluster does not exist
	run minikube stop
	run minikube delete

    minikube start

	IMAGE_NAME=kubernetes-grafana-test

    eval $(minikube docker-env)
    docker build .. -t $IMAGE_NAME

    kubectl delete clusterrolebinding --ignore-not-found=true $IMAGE_NAME
    kubectl delete deployment --ignore-not-found=true $IMAGE_NAME

    kubectl create clusterrolebinding $IMAGE_NAME --clusterrole=cluster-admin --serviceaccount=default:default
    kubectl run $IMAGE_NAME --image=$IMAGE_NAME --image-pull-policy=Never

    kubectl create -f crd.yaml

    kubectl apply -f grafana.yaml
}

teardownIntegrationTests() {
    minikube stop
    minikube delete
}

validateGrafanaUrl() {
    # get grafana url
    GRAFANA_URL=$(minikube service grafana --url --interval=1)

    echo "grafana url: " $GRAFANA_URL
    httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL})

    [ "$httpStatus" -eq "200" ]
}

#
# validateDashboardCount <count>
#   use grafana search api to confirm that the count is what is expected 
#
validateDashboardCount() {
    searchJson=$(curl --silent ${GRAFANA_URL}/api/search)

    count=$(echo $searchJson | jq length)

    [ "$count" -eq "$1" ]
}

#
# validatePostDashboard <yaml file name>
#   note that the dashboard file name must match the GrafanaDashboard object name ...
#    ... b/c i'm lazy
#
validatePostDashboard() {
    specfile=$1

    dashboardName="${specfile##*/}"
    dashboardName="${dashboardName%.*}"

    # create in kubernetes
    kubectl apply -f $specfile >&2

	sleep 5s

    dashboardId=$(kubectl get GrafanaDashboard -o=jsonpath="{.items[?(@.metadata.name==\"${dashboardName}\")].status.grafanaUID}")

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "200" ]

    echo $dashboardId
}

#
# validateDashboardContents <yaml file name>
#   creates a dashboard and both verifies it exists and that its content matches
#
validateDashboardContents() {
    filename=$1

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
}

#
# notification channel
#
#

# validatePostNotificationChannel <yaml file name>
#   note that the channel file name must match the GrafanaNotificationChannel object name ...
#    ... b/c i'm lazy
#
validatePostNotificationChannel() {
    specfile=$1

    channelName="${specfile##*/}"
    channelName="${channelName%.*}"

    # create in kubernetes
    kubectl apply -f $specfile >&2

	sleep 5s

    channelId=$(kubectl get GrafanaNotificationChannel -o=jsonpath="{.items[?(@.metadata.name==\"${channelName}\")].status.grafanaID}")

    [ "$channelId" != "" ]

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/alert-notifications/${channelId})
    [ "$httpStatus" -eq "200" ]

    echo $channelId
}


#
# validateNotificationChannelCount <count>
#   use grafana search api to confirm that the count is what is expected 
#
validateNotificationChannelCount() {
    channelJson=$(curl --silent ${GRAFANA_URL}/api/alert-notifications)

    count=$(echo $channelJson | jq length)

    [ "$count" -eq "$1" ]
}

#
# validateNotificationChannelContents <yaml file name>
#   creates a Notification Channel and both verifies it exists and that its content matches
#
validateNotificationChannelContents() {
    filename=$1

    channelId=$(validatePostNotificationChannel $filename)

    echo "Test Json Content of $filename ($channelId)"

    channelJsonFromGrafana=$(curl --silent ${GRAFANA_URL}/api/alert-notifications/${channelId})

    # grafana adds a lot of fields.  blindly stripping them off here.
    #   todo:  find a way to dynamically compare only the fields in json from the yaml spec file
    echo $channelJsonFromGrafana | jq 'del(.id) | del(.disableResolveMessage) | del(.created) | del(.frequency) | del(.updated)' > a.json

    channelJsonFromYaml=$(grep -A9999 'notificationChannelJson' $filename)
    channelJsonFromYaml=${channelJsonFromYaml%?}   # strip final quote
    channelJsonFromYaml=${channelJsonFromYaml#*\'} # strip up to and including the first quote

    echo $channelJsonFromYaml > b.json

    equal=$(jq --argfile a a.json --argfile b b.json -n '$a == $b')

    if [ "$equal" != "true" ]; then
        run diff <(jq -S . a.json) <(jq -S . b.json)
        echo $output
    fi

    [ "$equal" = "true" ]

    rm a.json
    rm b.json
}

#
# utils
#
dumpState() {
    kubectl logs $(kubectl get po -l=run=kubernetes-grafana-test --no-headers=true | cut -d ' ' -f 1)
    kubectl logs $(kubectl get po -l=app=grafana --no-headers=true| cut -d ' ' -f 1)

    kubectl describe GrafanaDashboard
    kubectl describe GrafanaNotificationChannel
}