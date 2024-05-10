#!/bin/bash
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
SA_NAME=${5:-quorum-sa}

echo "az get-credentials ..."
# if running this on a VM/Function/etc use a managed identity
# az login --identity --debug
# if running locally
az login
az aks get-credentials --resource-group "$AKS_RESOURCE_GROUP" --name "$AKS_CLUSTER_NAME"

# https://learn.microsoft.com/en-us/azure/aks/use-oidc-issuer
echo "Get the oidc issuer and workload identity ID from the cluster... "
AKS_MANAGED_IDENTITY_RESOURCE_ID=$(az identity show --name "$AKS_MANAGED_IDENTITY"  --resource-group "$AKS_RESOURCE_GROUP" | jq -r '.id')
AKS_OIDC_ISSUER=$(az aks show -n "$AKS_CLUSTER_NAME" -g "$AKS_RESOURCE_GROUP" --query "oidcIssuerProfile.issuerUrl" -otsv)

# https://learn.microsoft.com/en-gb/azure/aks/workload-identity-deploy-cluster#create-kubernetes-service-account
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    azure.workload.identity/client-id: "$AKS_MANAGED_IDENTITY_RESOURCE_ID"
  name: "$SA_NAME"
  namespace: "$AKS_NAMESPACE"
EOF

# Create the federated identity credential between the managed identity, the service account issuer, and the subject.
az identity federated-credential create --name "$AKS_MANAGED_IDENTITY-fc" --identity-name "$AKS_MANAGED_IDENTITY" --resource-group "$AKS_RESOURCE_GROUP" --issuer "$AKS_OIDC_ISSUER" --subject system:serviceaccount:"$AKS_NAMESPACE":"$SA_NAME" --audience api://AzureADTokenExchange

echo "Provisioning CSI drivers... "
# Helm charts for KeyVault drivers
helm repo add stable https://charts.helm.sh/stable
helm repo add csi-secrets-store-provider-azure https://azure.github.io/secrets-store-csi-driver-provider-azure/charts
helm repo update
helm upgrade --install --namespace "$AKS_NAMESPACE" --create-namespace akv-secrets-csi-driver csi-secrets-store-provider-azure/csi-secrets-store-provider-azure
helm ls --namespace "$AKS_NAMESPACE"

echo "Done... "
