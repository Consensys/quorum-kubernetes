# Helm Charts

Each helm chart that you can use has the following keys and you need to set them. The `cluster.provider` is used as a key for the various cloud features enabled. Also you only need to specify one cloud provider, **not** both if deploying to cloud. As of writing this doc, AWS and Azure are fully supported.

```bash
# dict with what features and the env you're deploying to
cluster:
  provider: local  # choose from: local | aws | azure
  cloudNativeServices: false # set to true to use Cloud Native Services (SecretsManager and IAM for AWS; KeyVault & Managed Identities for Azure)

aws:
  # the aws cli commands uses the name 'quorum-sa' so only change this if you altered the name
  serviceAccountName: quorum-sa
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
  # the subscription ID to use - this needs to be set explicitly when using multi tenancy
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
helm install elasticsearch --version 7.17.1 elastic/elasticsearch --namespace quorum --create-namespace --values ./values/elasticsearch.yml --set replicas=1 --set minimumMasterNodes=1
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

Additionally, you will need to deploy a separate ingress which will serve external facing services like the explorer and monitoring endpoints

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install quorum-monitoring-ingress ingress-nginx/ingress-nginx \
    --namespace quorum \
    --set controller.ingressClassResource.name="monitoring-nginx" \
    --set controller.ingressClassResource.controllerValue="k8s.io/monitoring-ingress-nginx" \
    --set controller.replicaCount=1 \
    --set controller.nodeSelector."kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."kubernetes\.io/os"=linux \
    --set controller.admissionWebhooks.patch.nodeSelector."kubernetes\.io/os"=linux \
    --set controller.service.externalTrafficPolicy=Local
```

Change Configmap of ingress nginx controller
```
kubectl edit -n quorum configmaps quorum-monitoring-ingress-ingress-nginx-controller
```
Then set `allow-snippet-annotations: "true"`
```
kubectl apply -f ../ingress/ingress-rules-monitoring.yml
```

Once complete, view the IP address listed under the `Ingress` section if you're using the Kubernetes Dashboard
or on the command line `kubectl -n quorum get services quorum-monitoring-ingress-ingress-nginx-controller`.

You can then access Grafana on: 
```bash
# For Besu's grafana address:
http://<INGRESS_IP>/d/XE4V0WGZz/besu-overview?orgId=1&refresh=10s

# For GoQuorum's grafana address:
http://<INGRESS_IP>/d/a1lVy7ycin9Yv/goquorum-overview?orgId=1&refresh=10s
```

You can access Kibana on:
```bash
http://<INGRESS_IP>/kibana
```

### Blockchain Explorer

#### Blockscout

```bash
helm dependency update ./charts/blockscout

# For GoQuorum
helm install blockscout ./charts/blockscout --namespace quorum --values ./values/blockscout-goquorum.yml

# For Besu
helm install blockscout ./charts/blockscout --namespace quorum --values ./values/blockscout-besu.yml
```
If you've deployed the Ingress from the previous step, you can access Blockscout on:

```bash
http://<INGRESS_IP>/blockscout
```

#### Quorum Explorer

You may optionally deploy our lightweight Quorum Explorer, which is compatible for both Besu and GoQuorum. The Explorer can give an overview over the whole network, such as querying each node on the network for block information, voting or removing validators from the network, demonstrating a SimpleStorage smart contract with privacy enabled, and sending transactions between wallets in one interface.

**Note:** It will be necessary to update the `quorum-explorer-config` configmap after deployment (if you do not choose to modify the `explorer-besu.yaml` or `explorer-goquorum.yaml` files before deploying) to provide the application endpoints to the nodes on the network. You may choose to either use internal k8s DNS or through ingress (your preference and needs). Please see the `values/explorer-besu.yaml` or `values/explorer-goquorum.yaml` to see some examples.

**Going into Production**

If you would like to use the Quorum Explorer in a production environment, it is highly recommended to enable OAuth or, at the minimum, local username and password authentication which is natively supported by the Explorer. The `explorerEnvConfig` section of the `explorer-besu.yaml` and `explorer-goquorum.yaml` files contain the variables that you may change. By default `DISABE_AUTH` is set to `true`, which means authentication is disabled. Change this to `false` if you would like to enable authentication. If this is set to `false`, you must also provide at least one authentication OAuth method by filling the variables below (supports all of those listed in the file).

You may find out more about the variables [here](https://github.com/ConsenSys/quorum-explorer#going-into-production).

```
To deploy for Besu:

```bash
helm install quorum-explorer ./charts/explorer --namespace quorum --create-namespace --values ./values/explorer-besu.yaml
```

To deploy for GoQuorum:
```bash
helm install quorum-explorer ./charts/explorer --namespace quorum --create-namespace --values ./values/explorer-goquorum.yaml

After modifying configmap with node details, you will need to restart the pod to get the config changes. Deleting the existing pod will force the deployment to recreate it:

```bash
kubectl delete pod <quorum-explorer-pod-name>
```

If you've deployed the Ingress from the previous step, you can access the Quorum Explorer on:

```bash
http://<INGRESS_IP>/explorer
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

Optionally deploy the ingress controller for the network and nodes like so:

NOTE: Deploying the ingress rules, assumes you are connecting to the `tx-1` node from section 3 above. Please update this as required to suit your requirements

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install quorum-network-ingress ingress-nginx/ingress-nginx \
    --namespace quorum \
    --set controller.ingressClassResource.name="network-nginx" \
    --set controller.ingressClassResource.controllerValue="k8s.io/network-ingress-nginx" \
    --set controller.replicaCount=1 \
    --set controller.nodeSelector."kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."kubernetes\.io/os"=linux \
    --set controller.admissionWebhooks.patch.nodeSelector."kubernetes\.io/os"=linux \
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

Optionally deploy the ingress controller for the network and nodes like so:

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install quorum-network-ingress ingress-nginx/ingress-nginx \
    --namespace quorum \
    --set controller.ingressClassResource.name="network-nginx" \
    --set controller.ingressClassResource.controllerValue="k8s.io/network-ingress-nginx" \
    --set controller.replicaCount=1 \
    --set controller.nodeSelector."kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."kubernetes\.io/os"=linux \
    --set controller.admissionWebhooks.patch.nodeSelector."kubernetes\.io/os"=linux \
    --set controller.service.externalTrafficPolicy=Local
```

Change Configmap of ingress nginx controller
```
kubectl edit -n quorum configmaps quorum-network-ingress-ingress-nginx-controller
```
Then set `allow-snippet-annotations: "true"`
```
kubectl apply -f ../ingress/ingress-rules-goquorum.yml
```

### Once deployed, services are available as follows on the IP/ of the ingress controllers:

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
