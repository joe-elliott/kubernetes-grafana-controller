#! /bin/bash

set -e
set -x

IMAGE_NAME=kubernetes-grafana-test

minikube status
eval $(minikube docker-env)

docker build .. -t $IMAGE_NAME

kubectl delete clusterrolebinding --ignore-not-found=true $IMAGE_NAME
kubectl delete deployment --ignore-not-found=true $IMAGE_NAME

kubectl create clusterrolebinding $IMAGE_NAME --clusterrole=cluster-admin --serviceaccount=default:default
kubectl run $IMAGE_NAME --image=$IMAGE_NAME --image-pull-policy=Never