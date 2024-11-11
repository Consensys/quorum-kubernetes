#!/usr/bin/env bash

set -e

rm -rf charts/besu-genesis
rm -rf charts/besu-node
rm -rf charts/blockscout

mkdir -p charts
cp -r ../../../helm/charts/besu-genesis charts/.
cp -r ../../../helm/charts/besu-node charts/.
cp -r ../../../helm/charts/blockscout charts/.

helm dependency update

helm upgrade besu-minikube --install --create-namespace --namespace besu-minikube .
