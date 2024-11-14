#!/usr/bin/env bash

set -e

helm upgrade genesis --install --namespace besu-minikube --create-namespace \
  --values deploy_flat_values.yaml ../../../helm/charts/besu-genesis
helm upgrade validator-1 --install --namespace besu-minikube ../../../helm/charts/besu-node
helm upgrade validator-2 --install --namespace besu-minikube ../../../helm/charts/besu-node
helm upgrade validator-3 --install --namespace besu-minikube ../../../helm/charts/besu-node
helm upgrade validator-4 --install --namespace besu-minikube ../../../helm/charts/besu-node
helm dependency update ../../../helm/charts/besu-node
helm upgrade blockscout  --install --namespace besu-minikube \
  --set blockscout.ethereum_jsonrpc_endpoint="besu-node-validator-1" \
  ../../../helm/charts/blockscout
kubectl apply --namespace besu-minikube -f templates/ingress.yaml
