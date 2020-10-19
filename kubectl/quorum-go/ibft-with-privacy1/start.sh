#!/bin/bash

echo "Checking if Kubernetes is running"
kubectl > /dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ];
then
  printf "Error: kubectl not found, please install kubectl before running this script.
"
  printf "For more information, see our qubernetes project: https://github.com/jpmorganchase/qubernetes
"
  exit $EXIT_CODE
fi

kubectl cluster-info > /dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ];
then
  printf "Could not connect to a kubernetes cluster. Please make sure you have minikube or another local kubernetes cluster running.
"
  printf "For more information, see our qubernetes project: https://github.com/jpmorganchase/qubernetes
"
  exit $EXIT_CODE
fi

echo "Setting up network"
kubectl apply -f out -f out/deployments
echo "
Run 'kubectl get pods' to check status of pods
"
