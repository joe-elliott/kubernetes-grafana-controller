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