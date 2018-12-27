export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false

SETUP_SUCCEEDED=0
GRAFANA_URL=""

setupIntegrationTests() {

    # ignore failure on these.  they will fail if a minikube cluster does not exist
    echo '# terminating minikube' >&3
	run minikube stop
	run minikube delete

    echo '# starting minikube' >&3
    minikube start

    echo '# building docker container' >&3
	IMAGE_NAME=kubernetes-grafana-test

    eval $(minikube docker-env)
    docker build .. -t $IMAGE_NAME

    echo '# configuring k8s' >&3
    kubectl delete clusterrolebinding --ignore-not-found=true $IMAGE_NAME
    kubectl delete deployment --ignore-not-found=true $IMAGE_NAME

    kubectl create clusterrolebinding $IMAGE_NAME --clusterrole=cluster-admin --serviceaccount=default:default
    kubectl run $IMAGE_NAME --image=$IMAGE_NAME --image-pull-policy=Never

    SETUP_SUCCEEDED=1
}

teardownIntegrationTests() {
    minikube stop
}

validateGrafanaUrl() {
    # get grafana url
    GRAFANA_URL=$(minikube service grafana --url)

    echo "grafana url: " $GRAFANA_URL
    httpStatus=$(curl --silent --output /dev/null --write-out "%{http_code}" ${GRAFANA_URL})

    [ "$httpStatus" -eq "200" ]
}

getGrafanaDashboardIdByName () {
    id=$(kubectl get GrafanaDashboard -o=jsonpath='{.items[?(@.metadata.name==\"${$1}\")].status.grafanaUID}')

    return $id
}