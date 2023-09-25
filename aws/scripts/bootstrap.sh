#!/bin/bash
#
# Run as:
# ./bootstrap.sh "AWS_REGION" "AWS_ACCOUNT" "CLUSTER_NAME" "AKS_NAMESPACE"
#

set -eux

AWS_REGION=${1:-rg}
AWS_ACCOUNT=${2:-account}
CLUSTER_NAME=${3:-cluster}
# quourum
AKS_NAMESPACE=${4:-quorum}

echo "aws get-credentials ..."
aws sts get-caller-identity
aws eks --region "${AWS_REGION}" update-kubeconfig --name "${CLUSTER_NAME}"

helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm install --namespace kube-system --create-namespace csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver
kubectl apply -f https://raw.githubusercontent.com/aws/secrets-store-csi-driver-provider-aws/main/deployment/aws-provider-installer.yaml 

# If you have deployed the above policy before, acquire its ARN:
POLICY_ARN=$(aws iam list-policies --scope Local --query 'Policies[?PolicyName==`quorum-node-secrets-mgr-policy`].Arn' --output text)
if [ $? -eq 1 ] 
then
  echo "Deploy the policy"
  POLICY_ARN=$(aws --region $AWS_REGION --query Policy.Arn --output text iam create-policy --policy-name quorum-node-secrets-mgr-policy --policy-document '{
      "Version": "2012-10-17",
      "Statement": [ {
          "Effect": "Allow",
          "Action": ["secretsmanager:CreateSecret","secretsmanager:UpdateSecret","secretsmanager:DescribeSecret","secretsmanager:GetSecretValue","secretsmanager:PutSecretValue","secretsmanager:ReplicateSecretToRegions","secretsmanager:TagResource"],
          "Resource": ["arn:aws:secretsmanager:$AWS_REGION:$AWS_ACCOUNT:secret:goquorum-node-*", "arn:aws:secretsmanager:$AWS_REGION:$AWS_ACCOUNT:secret:besu-node-*"]
      } ]
  }')
fi

eksctl create iamserviceaccount --name quorum-sa --namespace "${NAMESPACE}" --region="${AWS_REGION}" --cluster "${CLUSTER_NAME}" --attach-policy-arn "$POLICY_ARN" --approve --override-existing-serviceaccounts
echo "Done... "
