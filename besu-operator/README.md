# Kubernetes Operators


## CRD Structure

1. Besu
2. BesuNode
3. Grafana
4. Prometheus


## Deploying

1. kubectl apply -f deploy/crds/basiccrds/
2. kubectl apply -f deploy/crds/besu_without_keys.yaml
3. operator-sdk run local
