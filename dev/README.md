# Playground

The playground area has basic examples you can use to setup your network with. They use fixed keys and `kubectl` to deploy. Please refer to the individual folders to see how things are connected and working. You are encouraged to pull these manifests apart and experiment with options to learn how things work.

## Local Development:
- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/) This is the local equivalent of a K8S cluster
- [Helm](https://helm.sh/docs/)
- [Helmfile](https://github.com/roboll/helmfile)
- [Helm Diff plugin](https://github.com/databus23/helm-diff)

Minikube defaults to 2 CPU's and 2GB of memory, unless configured otherwise. We recommend you starting with at least 16GB, depending on the amount of nodes you are spinning up - the recommended requirements for each besu node are 4GB
```bash
minikube start --memory 16384
# or with RBAC
minikube start --memory 16384 --extra-config=apiserver.Authorization.Mode=RBAC
```

Verify kubectl is connected to Minikube with:
```bash
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.1", GitCommit:"4485c6f18cee9a5d3c3b4e523bd27972b1b53892", GitTreeState:"clean", BuildDate:"2019-07-18T09:18:22Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.0", GitCommit:"e8462b5b5dc2584fdcd18e6bcfe9f1e4d970a529", GitTreeState:"clean", BuildDate:"2019-06-19T16:32:14Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
```

## Usage

*For Besu:*
```bash

cd dev/charts/
helm install monitoring ./charts/besu-monitoring --namespace monitoring --create-namespace
helm install genesis ./charts/besu-genesis --namespace besu --create-namespace --values ./values/genesis-besu.yml 

helm install bootnode-1 ./charts/besu-node --namespace besu --values ./values/bootnode.yml
helm install bootnode-2 ./charts/besu-node --namespace besu --values ./values/bootnode.yml

helm install validator-1 ./charts/besu-node --namespace besu --values ./values/validator.yml
helm install validator-2 ./charts/besu-node --namespace besu --values ./values/validator.yml
helm install validator-3 ./charts/besu-node --namespace besu --values ./values/validator.yml
helm install validator-4 ./charts/besu-node --namespace besu --values ./values/validator.yml

# spin up a besu and orion node pair
helm install member-1 ./charts/besu-node --namespace besu --values ./values/txnode.yml
```

Optionally deploy the ingress controller like so:

NOTE: Deploying the ingress rules, assumes you are connecting to the `tx-1` node from section 3 above. Please update this as required to suit your requirements

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install besu-ingress ingress-nginx/ingress-nginx \
    --namespace besu \
    --set controller.replicaCount=2 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.admissionWebhooks.patch.nodeSelector."beta\.kubernetes\.io/os"=linux

kubectl apply -f ./ingress/ingress-rules-besu.yml
```

*For GoQuorum:*

Change directory to the charts folder ie `/charts/dev` or `/charts/prod`
```
cd helm/dev/  
# Please do not use this monitoring chart in prod, it needs authentication, pending close of https://github.com/ConsenSys/cakeshop/issues/86 
helm install monitoring ./charts/quorum-monitoring --create-namespace --namespace quorum 
helm install genesis ./charts/quorum-genesis --create-namespace --namespace quorum --values ./values/genesis-quorum.yml 

helm install validator-1 ./charts/quorum-node --namespace quorum --values ./values/validator.yml
helm install validator-2 ./charts/quorum-node --namespace quorum --values ./values/validator.yml
helm install validator-3 ./charts/quorum-node --namespace quorum --values ./values/validator.yml
helm install validator-4 ./charts/quorum-node --namespace quorum --values ./values/validator.yml

# spin up a quorum and tessera node pair
helm install tx-1 ./charts/quorum-node --namespace quorum --values ./values/txnode.yml
```

Optionally deploy the ingress controller like so:

NOTE: Deploying the ingress rules, assumes you are connecting to the `tx-1` node from section 3 above. Please update this as required to suit your requirements

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install quorum-ingress ingress-nginx/ingress-nginx \
    --namespace quorum \
    --set controller.replicaCount=2 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.admissionWebhooks.patch.nodeSelector."beta\.kubernetes\.io/os"=linux

kubectl apply -f ./ingress/ingress-rules-quorum.yml
```


4. Once deployed, services are available as follows on the IP/ of the ingress controllers:

Monitoring (if deployed)
```bash
# For Besu's grafana address: 
http://<INGRESS_IP>/d/XE4V0WGZz/besu-overview?orgId=1&refresh=10s

# For GoQuorum's cakeshop address: 
http://<INGRESS_IP>

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