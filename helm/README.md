# Helm Charts

Each helm chart that you can use has the following keys and you need to set them. The `cluster.provider` is used as a key for the various cloud features enabled. Also you need only specify one cloud provider, **not** both if deploying to cloud. As at writing this doc, AWS and Azure are fully supported.

```bash
# dict with what features and the env you're deploying to
cluster:
  provider: local  # choose from: local | aws | azure
  cloudNativeServices: false # set to true to use Cloud Native Services (SecretsManager and IAM for AWS; KeyVault & Managed Identities for Azure)

aws:
  # the aws cli commands uses the name 'quorum-node-secrets-sa' so only change this if you altered the name
  serviceAccountName: quorum-node-secrets-sa
  # the region you are deploying to
  region: ap-southeast-2

azure:
  # the script/bootstrap.sh uses the name 'quorum-pod-identity' so only change this if you altered the name
  identityName: quorum-pod-identity
  # the clientId of the user assigned managed identity created in the template
  identityClientId: azure-clientId
  keyvaultName: azure-keyvault
  # the tenant ID of the key vault
  tenantId: azure-tenantId
  # the subscription ID to use - this needs to be set explictly when using multi tenancy
  subscriptionId: azure-subscriptionId

```

Setting the `cluster.cloudNativeServices: true` will:

- Keys are stored in KeyVault or Secrets Manager
- We make use of Managed Identities or IAMs for access

You are encouraged to pull these charts apart and experiment with options to learn how things work.

## Local Development:

Minikube defaults to 2 CPU's and 2GB of memory, unless configured otherwise. We recommend you starting with at least 16GB, depending on the amount of nodes you are spinning up - the recommended requirements for each besu node are 4GB

```bash
minikube start --memory 16384 --cpus 2
# or with RBAC
minikube start --memory 16384 --cpus 2 --extra-config=apiserver.Authorization.Mode=RBAC

# enable the ingress
minikube addons enable ingress

# optionally start the dashboard
minikube dashboard &
```

Verify kubectl is connected to Minikube with: (please use the latest version of kubectl)

```bash
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.1", GitCommit:"4485c6f18cee9a5d3c3b4e523bd27972b1b53892", GitTreeState:"clean", BuildDate:"2019-07-18T09:18:22Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.0", GitCommit:"e8462b5b5dc2584fdcd18e6bcfe9f1e4d970a529", GitTreeState:"clean", BuildDate:"2019-06-19T16:32:14Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
```

## Usage

### _Spin up ELK for logs: (Optional but recommended)_

**NOTE:** this uses charts from Elastic - please configure this as per your requirements and policies

```bash
helm repo add elastic https://helm.elastic.co
helm repo update
# if on cloud
helm install elasticsearch --version 7.17.1 elastic/elasticsearch --namespace quorum --create-namespace --values ./values/elasticsearch.yml
# if local - set the replicas to 1
helm install elasticsearch --version 7.17.1 elastic/elasticsearch --namespace quorum --create-namespace --values ./values/elasticsearch.yml --set replicas=1 --set minimumMasterNodes: 1
helm install kibana --version 7.17.1 elastic/kibana --namespace quorum --values ./values/kibana.yml
helm install filebeat --version 7.17.1 elastic/filebeat  --namespace quorum --values ./values/filebeat.yml
```

Please also deploy the ingress (below) and the ingress rules to access kibana on path `http://<INGRESS_IP>/kibana`.
Alternatively configure the kibana ingress settings in the [values.yml](./values/kibana.yml)

Once you have kibana open, create a `filebeat` index pattern and logs should be available. Please configure this as
per your requirements and policies

### _Spin up prometheus-stack for metrics: (Optional but recommended)_

**NOTE:** this uses charts from prometheus-community - please configure this as per your requirements and policies

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
# NOTE: please refer to values/monitoring.yml to configure the alerts per your requirements ie slack, email etc
helm install monitoring prometheus-community/kube-prometheus-stack --version 34.10.0 --namespace=quorum --create-namespace --values ./values/monitoring.yml --wait
kubectl --namespace quorum apply -f  ./values/monitoring/
```

### _For Besu:_

```bash
helm install genesis ./charts/besu-genesis --namespace quorum --create-namespace --values ./values/genesis-besu.yml

# bootnodes - optional but recommended
helm install bootnode-1 ./charts/besu-node --namespace quorum --values ./values/bootnode.yml
helm install bootnode-2 ./charts/besu-node --namespace quorum --values ./values/bootnode.yml

# !! IMPORTANT !! - If you use bootnodes, please set `quorumFlags.usesBootnodes: true` in the override yaml files
# for validator.yml, txnode.yml, reader.yml
helm install validator-1 ./charts/besu-node --namespace quorum --values ./values/validator.yml
helm install validator-2 ./charts/besu-node --namespace quorum --values ./values/validator.yml
helm install validator-3 ./charts/besu-node --namespace quorum --values ./values/validator.yml
helm install validator-4 ./charts/besu-node --namespace quorum --values ./values/validator.yml

# spin up a besu and tessera node pair
helm install member-1 ./charts/besu-node --namespace quorum --values ./values/txnode.yml

# spin up a quorum rpc node
helm install rpc-1 ./charts/besu-node --namespace quorum --values ./values/reader.yml
```

Optionally deploy blockscout:

```bash
helm dependency update ./charts/blockscout
helm install blockscout ./charts/blockscout --namespace quorum --values ./values/blockscout-besu.yml
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

kubectl apply -f ../ingress/ingress-rules-besu.yml
```

### _For GoQuorum:_

```bash
helm install genesis ./charts/goquorum-genesis --namespace quorum --create-namespace --values ./values/genesis-goquorum.yml

helm install validator-1 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml
helm install validator-2 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml
helm install validator-3 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml
helm install validator-4 ./charts/goquorum-node --namespace quorum --values ./values/validator.yml

# spin up a quorum and tessera node pair
helm install member-1 ./charts/goquorum-node --namespace quorum --values ./values/txnode.yml

# spin up a quorum rpc node
helm install rpc-1 ./charts/goquorum-node --namespace quorum --values ./values/reader.yml
```

Optionally deploy blockscout:

```bash
helm dependency update ./charts/blockscout
helm install blockscout ./charts/blockscout --namespace quorum --values ./values/blockscout-goquorum.yml
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

kubectl apply -f ../ingress/ingress-rules-quorum.yml
```

### Once deployed, services are available as follows on the IP/ of the ingress controllers:

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

# which should return (confirming that the node running the JSON-RPC service is syncing):
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
