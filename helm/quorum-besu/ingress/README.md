
## Description:
This is in supplement to the helm roles and has not been added as a dependency of the Helm charts themselves to allow users to develop solutions that work for their needs.

We use this in our marketplace offerings on cloud providers like AWS, Azure, GKE etc to give users an overview of their private network.

## Usage:
The following will create an nginx ingress controller and rules that route public traffic to the grafana service internally.

1. Deploy the ingress controller like so in the `monitoring` namespace:
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add stable https://kubernetes-charts.storage.googleapis.com/
helm repo update
helm install grafana-ingress ingress-nginx/ingress-nginx \
    --namespace monitoring \
    --set controller.name=grafana-ingress \
    --set controller.watchNamespace=monitoring \
    --set controller.ingressClass=grafana \
    --set controller.replicaCount=2 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set rbac.create=false

```

Alternatively to install an ingress for the RPC node service:
Update the namespace to suit your deployment
```
helm install besu-ingress ingress-nginx/ingress-nginx \
    --namespace besu \
    --set controller.name=besu-ingress \
    --set controller.watchNamespace=besu \
    --set controller.ingressClass=besu \
    --set controller.replicaCount=2 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set rbac.create=false

```

2. Deploy the ingress rules like so:
For grafana:

```bash
kubectl apply -f ingress-rules-grafana.yml
```
For the Besu RPC service :

```bash
kubectl apply -f ingress-rules-besu.yml
```


