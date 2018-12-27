#! /bin/bash

set -e
set -x

IMAGE_NAME=kubernetes-grafana-test

eval $(minikube docker-env)

docker build .. -t $IMAGE_NAME

kubectl run $IMAGE_NAME --image=$IMAGE_NAME