#!/usr/bin/env bash
set -ex

#minikube start

IMAGE_NAME=kubernetes-grafana-test

eval $(minikube docker-env)
docker build .. -t $IMAGE_NAME

kubectl delete clusterrolebinding --ignore-not-found=true $IMAGE_NAME
kubectl delete deployment --ignore-not-found=true $IMAGE_NAME

kubectl create clusterrolebinding $IMAGE_NAME --clusterrole=cluster-admin --serviceaccount=default:default
kubectl run $IMAGE_NAME --image=$IMAGE_NAME --image-pull-policy=Never
kubectl expose deployment $IMAGE_NAME --port=80 --target-port=8080 --type=NodePort

kubectl delete --ignore-not-found=true -f crd.yaml
kubectl create -f crd.yaml

kubectl delete --ignore-not-found=true -f grafana.yaml
kubectl apply -f grafana.yaml
