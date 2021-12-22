
# Prod

The following will only work on Azure, please refer to the ARM templates deployment [documentation](./azure/README.md) before proceeding. Please proceed only after using the `dev` section to gain an understanding of how things work

The prod charts have:
- Dynamic key & account generation
- Can be used only in cloud. Edit the `provider: azure` value in the values.yml.
- Keys are stored in keyvault on Azure and pods make use of Managed Identities
- To connect from your local machine you will need to use an ingress controller

You are encouraged to use a HA SQL database for Tessera nodes (create externally and pass connection details to the values.yml)

## Development:
- [Helm](https://helm.sh/docs/)
- [Helmfile](https://github.com/roboll/helmfile)
- [Helm Diff plugin](https://github.com/databus23/helm-diff)


## Usage

Ensure you have followed the steps completed to setup the clusters in (AWS)[../aws/README.md] or in (Azure)[../azure/README.md]. The Azure setup requires a script to be run to provision drivers on the cluster

Once completed, update the `aws` or `azure` section of the [values yml files](./helm/values/)  

Verify kubectl is connected to your cluster with:
```bash
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.1", GitCommit:"4485c6f18cee9a5d3c3b4e523bd27972b1b53892", GitTreeState:"clean", BuildDate:"2019-07-18T09:18:22Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.0", GitCommit:"e8462b5b5dc2584fdcd18e6bcfe9f1e4d970a529", GitTreeState:"clean", BuildDate:"2019-06-19T16:32:14Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
```

Then deploy the charts like so: 

*For Besu:*
```bash

cd dev/helm/
helm install monitoring ./charts/quorum-monitoring --namespace quorum --create-namespace
helm install genesis ./charts/besu-genesis --namespace quorum --create-namespace --values ./values/genesis-besu.yml

helm install bootnode-1 ./charts/besu-node --namespace quorum --values ./values/bootnode.yml
helm install bootnode-2 ./charts/besu-node --namespace quorum --values ./values/bootnode.yml

helm install validator-1 ./charts/besu-node --namespace quorum --values ./values/validator.yml
helm install validator-2 ./charts/besu-node --namespace quorum --values ./values/validator.yml
helm install validator-3 ./charts/besu-node --namespace quorum --values ./values/validator.yml
helm install validator-4 ./charts/besu-node --namespace quorum --values ./values/validator.yml

# spin up a besu and orion node pair
helm install member-1 ./charts/besu-node --namespace quorum --values ./values/txnode.yml
```

Optionally deploy the ingress controller like so:

NOTE: Deploying the ingress rules, assumes you are connecting to the `tx-1` node from section 3 above. Please update this as required to suit your requirements

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install besu-ingress ingress-nginx/ingress-nginx \
    --namespace quorum \
    --set controller.replicaCount=1 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.admissionWebhooks.patch.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.service.externalTrafficPolicy=Local

kubectl apply -f ../../ingress/ingress-rules-besu.yml
```

*For GoQuorum:*
```bash
cd dev/helm/  
helm install monitoring ./charts/quorum-monitoring --namespace quorum --create-namespace
helm install genesis ./charts/goquorum-genesis --namespace quorum --create-namespace --values ./values/genesis-goquorum.yml

helm install validator-1 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml
helm install validator-2 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml
helm install validator-3 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml
helm install validator-4 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml

# spin up a quorum and tessera node pair
helm install member-1 ./charts/goquorum-node --namespace quorum --values ./values/txnode.yml
```

Optionally deploy the ingress controller like so:

NOTE: Deploying the ingress rules, assumes you are connecting to the `tx-1` node from section 3 above. Please update this as required to suit your requirements

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install quorum-ingress ingress-nginx/ingress-nginx \
    --namespace quorum \
    --set controller.replicaCount=1 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.admissionWebhooks.patch.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.service.externalTrafficPolicy=Local

kubectl apply -f ../../ingress/ingress-rules-quorum.yml
```


4. Once deployed, services are available as follows on the IP/ of the ingress controllers:

Monitoring (if deployed)
```bash
# For Besu's grafana address:
http://<INGRESS_IP>/d/XE4V0WGZz/besu-overview?orgId=1&refresh=10s

# For GoQuorum's grafana address:
http://<INGRESS_IP>/d/a1lVy7ycin9Yv/goquorum-overview?orgId=1&refresh=10s

```


API Calls to either client
```bash

# HTTP RPC API:
curl -v -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://<INGRESS_IP>/rpc

# which should return (confirming that the node running the JSON-RPC service has peers):
{
  "jsonrpc" : "2.0",
  "id" : 1,
  "result" : "0x4e9"
}

# HTTP GRAPHQL API:
curl -X POST -H "Content-Type: application/json" --data '{ "query": "{syncing{startingBlock currentBlock highestBlock}}"}' http://<INGRESS_IP>/graphql/
# which should return
{
  "data" : {
    "syncing" : null
  }
}
```
