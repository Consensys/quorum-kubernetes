# Besu Enterprise Ethereum Client

[Besu](https://besu.hyperledger.org/en/latest/) is an open-source [Ethereum](https://www.ethereum.org/) client developed under the Apache 2.0 license and written in Java. 
It runs on the Ethereum public network, private networks, and test networks such as Rinkeby, Ropsten, and Görli. 
Besu implements Proof of Work (Ethash) and Proof of Authority ([IBFT 2.0](https://besu.hyperledger.org/en/latest/Consensus-Protocols/IBFT/) and [Clique](https://besu.hyperledger.org/en/latest/Consensus-Protocols/Clique/)) consensus mechanisms.

You can use Besu to develop enterprise applications requiring secure, high-performance transaction processing in a private network.

Besu supports enterprise features including privacy and permissioning.

## Introduction
This chart deploys a private Ethereum network with PoA (IBFT 2.0) consensus onto a Kubernetes cluster using `Helm Chart` and `helmfile`. 
For further information on running a private network, refer to [Besu's documentation](https://besu.hyperledger.org/en/latest/). 

In IBFT 2.0 networks, transactions and blocks are validated by approved accounts, known as validators. 
Validators take turns to create the next block. Existing validators propose and vote to add or remove validators.

Minimum Number of Validators IBFT 2.0 requires **4 validators** to be Byzantine fault tolerant.

This charts deploys 3 components:
* genesis: generates the genesis file and validator keys and bootnode keys
* node: the basic node 
* monitoring: prometheus & grafana monitoring setup with the custom Besu dashboard

## Prerequisites
- [Kubernetes](https://kubernetes.io/) 1.12+
- [Helm](https://helm.sh/docs/)
- [helmfile](https://github.com/roboll/helmfile)
- [Helm Diff plugin](https://github.com/databus23/helm-diff)


## Known Limitations
There is currently a limitation in that using this chart with separate bootnodes on a local cluster like Minikube or MicroK8S has issues syncing because of the way the routing is handled.
This is not the case on full clusters or in cloud. On local setups we recommend using Validators 1 & 2 as bootnodes 

## Installing the Chart
To install the chart in the namesapce with the name `my-namespace`:
```bash
helmfile -n my-namespace -f helmfile.yaml apply
```


The command deploys multi Besu nodes in PoA (IBFT 2.0) on the Kubernetes cluster in the default configuration. 
The configuration section lists the parameters that can be configured during installation.

> **Tip**: If there are problems to deploy, update your `Helm` and your `helmfile`

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart
To uninstall/delete the deployment:

`helmfile -n my-namespace -f helmfile.yaml delete --purge`

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

Please have a look at [node-values.yaml](charts/besu-node/values.yaml) and [genesis-values.yaml](charts/besu-genesis/values.yaml) 
to show all parameters.

The [genesis-values.yaml](charts/besu-genesis/values.yaml) contains parameters for the genesis file creation, validator & bootnode keys  

The following table lists the configurable parameters of the **Besu genesis** chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`image.repository` | Container image repository | `hyperledger/besu`
`image.tag` | Container image tag | `1.1.3`
`image.pullPolicy` | Container image pull policy | `IfNotPresent`
`rawGenesisConfig.genesis.blockchain.nodes.generate` | If true, generate the number of key-pairs | `true`
`rawGenesisConfig.genesis.blockchain.nodes.count` | The number of key-pairs generated for validators and inject into extraData | `4`
`rawGenesisConfig.genesis.blockchain.nodes.keys` | The list of private key to inject into extraData | `none`
`rawGenesisConfig.genesis.config.chainId` | The identifier of the private Ethereum network | `1981`
`rawGenesisConfig.genesis.config.constantinoplefixblock` | In private networks, the milestone block defines the protocol version for the network | `0`
`rawGenesisConfig.genesis.config.ibft2.blockperiodseconds` | Minimum block time in seconds. | `2`
`rawGenesisConfig.genesis.config.ibft2.epochlength` | Number of blocks after which to reset all votes. | `30000`
`rawGenesisConfig.genesis.config.ibft2.requesttimeoutseconds` | Timeout for each consensus round before a round change. | `10`
`rawGenesisConfig.genesis.config.extraData` | The extraData property is RLP encoded. | `0x`
`rawGenesisConfig.genesis.config.nonce` |  | `0x0`
`rawGenesisConfig.genesis.config.timestamp` |  | `0x58ee40ba`
`rawGenesisConfig.genesis.config.gasLimit` | Set the block size limit (measured in gas) | `0x47b760`
`rawGenesisConfig.genesis.config.difficulty` | Specify a fixed difficulty in private networks | `0x0`
`rawGenesisConfig.genesis.config.mixHash` | Hash for Istanbul block identification (IBFT 2.0). | `0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365`
`rawGenesisConfig.genesis.config.coinbase` | The coinbase address is the account to which mining rewards are paid. | `0x0000000000000000000000000000000000000000`
`rawGenesisConfig.genesis.config.alloc` | Predeploy contracts when starting Besu with Ether | `{config.genesis.config.alloc}`

For the other parameters, see [genesis-values.yaml](charts/besu-genesis/values.yaml)


The following table lists the configurable parameters of the **Besu node** chart and their default values.
Parameter | Description | Default
--------- | ----------- | -------
`image.repository` | Container image repository | `hyperledger/besu`
`image.tag` | Container image tag | `1.1.3`
`image.pullPolicy` | Container image pull policy | `IfNotPresent`
`besu.genesis.name` | IMPORTANT: The name of the configMap to retrieve the genesis | `genesis-besu`
`besu.bootnode.enabled` | If true, the Besu node deployed will be a bootnode | `false`
`besu.bootnode.privKey` | the Besu bootnode private key. If not present, the key is automatically generated | ``
`besu.validators.enabled` | If true, the Besu node deployed will be a validator | `false`
`besu.validators.privKey` | the Besu validator private key. If not present, retrieve the key from genesis chart  | ``
`replicaCount` | Warning: Should stay at this default value.  | `1`
`service.type` | Kubernetes service type | `ClusterIP`
`besu.persistentVolume.enabled` | If true, it's claim a persistent Volume | `false`
`besu.persistentVolume.size` | Size of the Volume | `2Gi`
`besu.persistentVolume.storageClass` | Storage class of the Volume | ``

For the other default parameters, see [node-values.yaml](charts/besu-node/values.yaml)


### Modify the number of validators
To modify the number of validators, you need to change values in two places.

Into the file `helmfile.yaml`, set (copy/past) the release section. 
You have to modify the `name` and the `index` value:

```yaml
  - name: validator-<INDEX_NUMBER>
    labels:
      component: validators
    namespace: {{ .Release.Namespace }}
    chart: ./besu-node
    values:
      - ./values/validator.yaml

```

Into the file `values/genesis.yaml`, change the `count` number to specify how many validators you want in the `genesis` file:

```yaml
config:
  blockchain:
    nodes:
      count: <NUMBER_OF_VALIDATORS>
```
