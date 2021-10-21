#!/bin/bash
#
# This bootstraps the ops vm to run helm charts on the aks cluster.
# This is for dev only at present, and this functionality will be moved to a lambda function
#
# Run as:
# ./bootstrap.sh "AKS_RESOURCE_GROUP" "AKS_CLUSTER_NAME" "AKS_MANAGED_IDENTITY" "AKS_NAMESPACE"
#
set -eux

AKS_RESOURCE_GROUP=${1:-rg}
AKS_CLUSTER_NAME=${2:-cluster}
AKS_MANAGED_IDENTITY=${3:-identity}
# quourum
AKS_NAMESPACE=${4:-quorum}

echo "az get-credentials ..."
# if running this on a VM/Function/etc use a managed identity
# az login --identity --debug
# if running locally
az login

# The pod identity cant be done via an ARM template and can only be done via CLI, hence
# https://docs.microsoft.com/en-us/azure/aks/use-azure-ad-pod-identity
echo "Update the cluster to use pod identity ... "
az aks update --name "$AKS_CLUSTER_NAME"  --resource-group "$AKS_RESOURCE_GROUP" --enable-pod-identity

echo "Provisioning AAD pod-identity... "
AKS_MANAGED_IDENTITY_RESOURCE_ID=$(az identity show --name "$AKS_MANAGED_IDENTITY"  --resource-group "$AKS_RESOURCE_GROUP" | jq -r '.id')
az aks pod-identity add \
    --resource-group "$AKS_RESOURCE_GROUP" \
    --cluster-name "$AKS_CLUSTER_NAME" \
    --identity-resource-id "$AKS_MANAGED_IDENTITY_RESOURCE_ID" \
    --namespace "$AKS_NAMESPACE" \
    --name quorum-pod-identity >/dev/null


echo "Provisioning CSI drivers... "
az aks get-credentials --resource-group "$AKS_RESOURCE_GROUP" --name "$AKS_CLUSTER_NAME" --admin
# Helm charts for KeyVault drivers
helm repo add stable https://charts.helm.sh/stable
helm repo add csi-secrets-store-provider-azure https://raw.githubusercontent.com/Azure/secrets-store-csi-driver-provider-azure/master/charts
helm repo update
helm upgrade --install --namespace "$AKS_NAMESPACE" --create-namespace akv-secrets-csi-driver csi-secrets-store-provider-azure/csi-secrets-store-provider-azure
helm ls --namespace "$AKS_NAMESPACE"

echo "Done... "
