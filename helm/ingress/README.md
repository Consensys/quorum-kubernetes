
## Description:
This is in supplement to the helm roles and has not been added as a dependency of the Helm charts themselves to allow users to develop solutions that work for their needs.

We use this in our marketplace offerings on cloud providers like AWS, Azure, GKE etc to give users an overview of their private network.

## Usage:
The following will create an nginx ingress controller and rules that route public traffic to the grafana service internally.

1. Deploy the ingress controller like so in the `monitoring` namespace:
```bash
helm repo add stable https://kubernetes-charts.storage.googleapis.com/
helm install grafana-ingress stable/nginx-ingress --namespace monitoring --set controller.replicaCount=2 --set rbac.create=true
``````

2. Deploy the ingress rules for grafana like so:
```bash
kubectl apply -f ingress-rules-grafana.yml
```


