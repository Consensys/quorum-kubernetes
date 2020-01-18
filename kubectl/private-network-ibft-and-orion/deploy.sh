kubectl apply -f namespace/monitoring-namespace.yaml
kubectl apply -f rbac/prometheus-rbac.yaml
kubectl apply -f secrets/validator-keys-secret.yaml
kubectl apply -f secrets/orion-keys-secret.yaml
kubectl apply -f configmap/configmap.yaml
kubectl apply -f configmap/orion-configmap.yaml
kubectl apply -f configmap/prometheus-configmap.yaml
kubectl apply -f configmap/grafana-configmap.yaml

kubectl apply -f deployments/orion1-deployment.yaml
kubectl apply -f deployments/orion2-deployment.yaml
kubectl apply -f deployments/orion3-deployment.yaml
kubectl apply -f deployments/orion4-deployment.yaml
kubectl apply -f deployments/validator1-deployment.yaml
kubectl apply -f deployments/validator2-deployment.yaml
kubectl apply -f deployments/validator3-deployment.yaml
kubectl apply -f deployments/validator4-deployment.yaml
kubectl apply -f deployments/node-deployment.yaml
kubectl apply -f deployments/prometheus-deployment.yaml
kubectl apply -f deployments/grafana-deployment.yaml

kubectl apply -f services/validator1-service.yaml
kubectl apply -f services/validator2-service.yaml
kubectl apply -f services/validator3-service.yaml
kubectl apply -f services/validator4-service.yaml
kubectl apply -f services/node-service.yaml
kubectl apply -f services/orion1-service.yaml
kubectl apply -f services/orion2-service.yaml
kubectl apply -f services/orion3-service.yaml
kubectl apply -f services/orion4-service.yaml
kubectl apply -f services/prometheus-service.yaml
kubectl apply -f services/grafana-service.yaml


