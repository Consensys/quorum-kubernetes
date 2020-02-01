kubectl apply -f namespace/

kubectl apply -f deployments/bootnode1-deployment.yaml
kubectl apply -f deployments/bootnode2-deployment.yaml
kubectl apply -f deployments/node-deployment.yaml
kubectl apply -f deployments/prometheus-deployment.yaml
kubectl apply -f deployments/grafana-deployment.yaml

kubectl apply -f rbac/
kubectl apply -f secrets/
kubectl apply -f configmap/
kubectl apply -f services/
