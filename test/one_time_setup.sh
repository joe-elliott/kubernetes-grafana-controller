#!/usr/bin/env bash

#minikube stop
#minikube delete

set -ex

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
