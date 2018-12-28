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
}

teardownIntegrationTests() {
    minikube stop
    minikube delete
}

validateGrafanaUrl() {
    # get grafana url
    GRAFANA_URL=$(minikube service grafana --url)

    echo "grafana url: " $GRAFANA_URL
    httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL})

    [ "$httpStatus" -eq "200" ]
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

    echo "Posting dashboard " $dashboardName

    # create in kubernetes
    kubectl apply -f $specfile

	sleep 5s

    dashboardId=$(kubectl get GrafanaDashboard -o=jsonpath="{.items[?(@.metadata.name==\"${dashboardName}\")].status.grafanaUID}")

    echo "Grafana Dashboard Id " $dashboardId

    # check if exists in grafana
	httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL}/api/dashboards/uid/${dashboardId})
    [ "$httpStatus" -eq "200" ]
}