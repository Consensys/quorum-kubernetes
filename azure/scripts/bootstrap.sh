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
SA_NAME=${5:-quorum}

echo "az get-credentials ..."
# if running this on a VM/Function/etc use a managed identity
# az login --identity --debug
# if running locally
az login

# https://learn.microsoft.com/en-us/azure/aks/use-oidc-issuer
echo "Update the cluster to use oidc issuer and workload identity ... "
az aks update -g myResourceGroup -n myAKSCluster --enable-oidc-issuer --enable-workload-identity

echo "Provisioning AAD pod-identity... "
AKS_MANAGED_IDENTITY_RESOURCE_ID=$(az identity show --name "$AKS_MANAGED_IDENTITY"  --resource-group "$AKS_RESOURCE_GROUP" | jq -r '.id')
AKS_OIDC_ISSUER=$(az aks show --name "$AKS_MANAGED_IDENTITY" --resource-group "$AKS_RESOURCE_GROUP"     --query "oidcIssuerProfile.issuerUrl" -otsv)

# https://learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    azure.workload.identity/client-id: "${AKS_MANAGED_IDENTITY_RESOURCE_ID}"
  name: "${SA_NAME}"
  namespace: "${AKS_NAMESPACE}"
EOF

cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: "${AKS_NAMESPACE}"
  name: "${SA_NAME}"
rules:
- apiGroups: [""]
  resources: ["secrets", "configmaps"]
  verbs: ["create", "get", "list", "update", "delete", "patch"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list"]
EOF  

cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: "${SA_NAME}"
  namespace: "${AKS_NAMESPACE}"
subjects:
- kind: ServiceAccount
  name: "${SA_NAME}"
  namespace: "${AKS_NAMESPACE}"
roleRef:
  kind: Role
  name: "${SA_NAME}"
  apiGroup: rbac.authorization.k8s.io
EOF   

az identity federated-credential create --name aks-federated-credential --identity-name "${AKS_MANAGED_IDENTITY}" --resource-group "${RESOURCE_GROUP}" --issuer "${AKS_OIDC_ISSUER}" --subject system:serviceaccount:"${AKS_NAMESPACE}":"${SA_NAME}" --audience api://AzureADTokenExchange



echo "Provisioning CSI drivers... "
az aks get-credentials --resource-group "$AKS_RESOURCE_GROUP" --name "$AKS_CLUSTER_NAME" --admin
# Helm charts for KeyVault drivers
helm repo add stable https://charts.helm.sh/stable
helm repo add csi-secrets-store-provider-azure https://raw.githubusercontent.com/Azure/secrets-store-csi-driver-provider-azure/master/charts
helm repo update
helm upgrade --install --namespace "$AKS_NAMESPACE" --create-namespace akv-secrets-csi-driver csi-secrets-store-provider-azure/csi-secrets-store-provider-azure
helm ls --namespace "$AKS_NAMESPACE"

echo "Done... "
