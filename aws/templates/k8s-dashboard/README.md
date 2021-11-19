# Install the kubernetes dashboard - only do this on dev clusters

Based on: https://docs.aws.amazon.com/eks/latest/userguide/dashboard-tutorial.html

### 1. Deploy the dash

```bash
# deploy the dashboard
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.5/aio/deploy/recommended.yaml
kubectl apply -f eks-admin-service-account.yml
kubectl proxy &

# get the token
kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep eks-admin | awk '{print $1}') | grep token: | awk '{print $2}'
```

### 3.Open a tab and go to
Go to:
```js
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#!/login
```

Paste the token in from step 1.