#!/usr/bin/env bash

set -x

rm -rf charts/besu-genesis
rm -rf charts/besu-node
rm -rf charts/blockscout

cp -rf ../../../helm/charts/besu-genesis charts/.
cp -rf ../../../helm/charts/besu-node charts/.
cp -rf ../../../helm/charts/blockscout charts/.

helm dependency update

helm upgrade besu-minikube --install --create-namespace --namespace besu-minikube .
