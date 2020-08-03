# Kubernetes Operators

[Docker Hub Repository](https://hub.docker.com/repository/docker/hyperledgerbesu/operators)

## Besu Configuration Options

| Field                  | Type                                | Description                                                                             |
| ---------------------- | ----------------------------------- | --------------------------------------------------------------------------------------- |
| BootnodesCount         | Integer                             | Number of bootnodes in the besu network                                                 |
| ValidatorsCount        | Integer                             | Number of validators in the besu network                                                |
| Members                | Integer                             | Number of member nodes in the besu network                                              |
| BootnodeKeys           | Key(PrivKey,PubKey)                 | Optional field to specify bootnodes keys, if not specified operator will generate keys  |
| ValidatorKeys          | Key(PrivKey,PubKey)                 | Optional field to specify validators keys, if not specified operator will generate keys |
| BesuNodeSpec           | BesuNodeSpec(Image,Resources,etc)   | Optional field to specify common besu node configuration                                |
| GenesisJSON            | GenesisJSON                         | Configuration for genesis block                                                         |
| BootnodesAreValidators | Boolean                             | Specifies whether bootnodes are validators or not, by default set to false              |
| Monitoring             | Boolean                             | Specifies whether to deploy grafana/prometheus or not                                   |
| PrometheusSpec         | PrometheusSpec(Image,Resources,etc) | Prometheus configuration options                                                        |
| GrafanaSpec            | GrafanaSpec(Image,Resources,etc)    | Grafana configuration options                                                           |

## Running operator as docker image

### Prerequisites

1. kubectl version v1.11.3+.
2. Access to a Kubernetes v1.11.3+ cluster.

### Steps

1. `kubectl apply -f deploy/service_account.yaml`
2. `kubectl apply -f deploy/role.yaml`
3. `kubectl apply -f deploy/role_binding.yaml`
4. `kubectl apply -f deploy/crds/basiccrds/`
5. `kubectl apply -f deploy/operator.yaml`
6. `kubectl apply -f deploy/crds/besu_without_keys.yaml`

## Running operator locally

### Prerequisites

1. go version v1.13+.
2. docker version 17.03+.
3. kubectl version v1.11.3+.
4. kustomize v3.1.0+
5. Access to a Kubernetes v1.11.3+ cluster.

### Steps

1. `kubectl apply -f deploy/crds/basiccrds/`
2. `kubectl apply -f deploy/crds/besu_without_keys.yaml`
3. `operator-sdk run local`

## Sample Configurations

1.  Without keys :

        apiVersion: hyperledger.org/v1alpha1
        kind: Besu
        metadata:
          name: besu
        spec:
          bootnodescount: 2
          validatorscount: 2
          members: 1

2.  With keys :

- If user provides more keys than required, then first m or n ( bootnodes or validators count ) keys will be used
- If user provides less keys than required, then new keys will be generated

        apiVersion: hyperledger.org/v1alpha1
        kind: Besu
        metadata:
          name: besu
        spec:
          bootnodescount: 2
          validatorscount: 2
          bootnodeKeys:
            - pubkey: "5d812c3c25ff398ab416968fce9009c2be7ed70a87abc8ea30bd667ce17a9287a6341fbf6ce757bb8148436c39c71296639ea81afcc94cdf908b6e1344f26188"
              privkey: "0x8457b1dc606d05308cc96f604dbde54ad85cb4de508742c8e265080b9e08b48c"
            - pubkey: "b2fba529681ea7f4619556753d40c8689b936fb1c621bc91f94d2938eb58c285d4911457ae4887b9c3bd593b2d608d319c6dc384d6acae2d043a4657029178d3"
              privkey: "0xa53a0dea5e51a9e0735e331b336af65d60f516c7080a52da683f6fd96342be42"
          validatorKeys:
            - pubkey: "00b20ab6a385a2403d64637b3d93cb6d83215a08f29adb6feb4b8bf03387b734444e8b060f53150dea4b9b897823540d19918c13d6f57a5153d190b5fad7bf51"
              privkey: "0x28021485044bf6870160f90b82ab6bbd14c691d4414140a63d047a7e3362c64e"
            - pubkey: "5fc1f8dc9f0c03087128e4bd724530e883d7de1a431269876dff9c95b8952f73c7e85ac7b49d85a2ad4950e967319482af435e07a0eab0a98d98449437787a00"
              privkey: "0xfb122a05ab1897ff144e1c9efb0bb3144f1e7f319aa5c55e20ef3d8d8464f4e8"
          members: 1

3. Bootnodes are validators or not

- By default, bootnodes will be distinct from validators, `bootnodesarevalidators` flag can be set to true to make sure bootnodes are also validators

        apiVersion: hyperledger.org/v1alpha1
        kind: Besu
        metadata:
          name: besu
        spec:
          bootnodescount: 2
          validatorscount: 2
          members: 1
          bootnodesarevalidators: true

4. With genesis configurations

- By default these values will be taken for `genesis.json`, this can be changed by specifying different configuration options in the following way

        apiVersion: hyperledger.org/v1alpha1
        kind: Besu
        metadata:
            name: besu
        spec:
            bootnodescount: 2
            validatorscount: 4
            members: 2
            genesis.json:
                genesis:
                config:
                    chainId: 2018
                    constantinoplefixblock: 0
                    ibft2:
                    blockperiodseconds: 2
                    epochlength: 30000
                    requesttimeoutseconds: 10
                nonce: "0x0"
                timestamp: "0x58ee40ba"
                gasLimit: "0x1fffffffffffff"
                difficulty: "0x1"
                mixHash: "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365"
                coinbase: "0x0000000000000000000000000000000000000000"
                alloc:
                    "9811ebc35d7b06b3fa8dc5809a1f9c52751e1deb":
                    balance: "0xad78ebc5ac6200000"

5. With besunode configurations

- By default following configuration values will be considered, this can be changed, same node configuration will apply to all nodes

        apiVersion: hyperledger.org/v1alpha1
        kind: Besu
        metadata:
            name: besu
        spec:
            bootnodescount: 2
            validatorscount: 2
            members: 1
            besunodespec:
                replicas: 2
                image:
                    pullPolicy: IfNotPresent
                    repository: hyperledger/besu
                    tag: 1.4.6
                resources:
                    cpuLimit: 500m
                    cpuRequest: 100m
                    memLimit: 2048Mi
                    memRequest: 1024Mi
                graphql:
                    authenticationEnabled: false
                    enabled: false
                    host: 0.0.0.0
                    port: 8547
                metrics:
                    enabled: true
                    host: 0.0.0.0
                    port: 9545
                p2p:
                    authenticationEnabled: false
                    discovery: true
                    enabled: true
                    host: 0.0.0.0
                    port: 30303
                rpc:
                    authenticationEnabled: false
                    enabled: true
                    host: 0.0.0.0
                    port: 8545
                ws:
                    authenticationEnabled: false
                    enabled: false
                    host: 0.0.0.0
                    port: 8546

6. With monitoring configurations

- By default monitoring will be set to true and following default values will be considered, they can be configured like below

        apiVersion: hyperledger.org/v1alpha1
        kind: Besu
        metadata:
            name: besu
        spec:
            bootnodescount: 2
            validatorscount: 2
            members: 1
            monitoring: true
            prometheusspec:
                image:
                pullPolicy: IfNotPresent
                repository: prom/prometheus
                tag: v2.11.1
                nodeport: 30090
                replicas: 1
                resources:
                cpuLimit: 500m
                cpuRequest: 100m
                memLimit: 512Mi
                memRequest: 256Mi
            grafanaspec:
                image:
                pullPolicy: IfNotPresent
                repository: grafana/grafana
                tag: 6.2.5
                nodeport: 30030
                replicas: 1
                resources:
                cpuLimit: 500m
                cpuRequest: 100m
                memLimit: 512Mi
                memRequest: 256Mi

## Version upgrade

- Supports upgraging besu version or changing replicas or changing number of member nodes
- `kubectl edit besu <besu_name>` :
  - For changing besu image, change spec:besunodespec:image
  - For changing member nodes, change spec:members
  - For changing replicas of validator or bootnodes, change spec:besunodespec:replicas

## Development

### Directory layout

      build/      # Contains built files & docker image
      deploy/
        crds/     # Contains Custom Resources definitions and different examples
        operator.yaml # Deployment corresponding to operator
        role.yaml   # Role required for operator
        role_binding.yaml  # Role binding required for operator
        service_account.yaml # Service account required for operator
      pkd/
        apis/
            hyperledger/
                v1alpha1/
                    besu_types.go  # Different types/structs required for besu CRD
                    besunode_types.go  # Different types/structs required for besunode CRD
                    grafana_types.go  # Different types/structs required for grafana CRD
                    prometheus_types.go # Different types/structs required for prometheus CRD
        controller/
            besu/
                besu_controller.go # Main reconcile controller logic
                besu_ensure.go     # Ensures existence & correct config of different child resources
                besu_resources.go  # Templates for new child resources
            besunode/
                besunode_controller.go
                besunode_ensure.go
                besunode_resources.go
            grafana/
                grafana_controller.go
                grafana_ensure.go
                grafana_resources.go
            prometheus/
                prometheus_controller.go
                prometheus_ensure.go
                prometheus_resources.go
        resources/
            common.go  # Common functions required across different controllers

### CRD Structure

1. Besu

   Spec Fields

| Field                  | Type                                | Description                                                                             |
| ---------------------- | ----------------------------------- | --------------------------------------------------------------------------------------- |
| BootnodesCount         | Integer                             | Number of bootnodes in the besu network                                                 |
| ValidatorsCount        | Integer                             | Number of validators in the besu network                                                |
| Members                | Integer                             | Number of member nodes in the besu network                                              |
| BootnodeKeys           | Key(PrivKey,PubKey)                 | Optional field to specify bootnodes keys, if not specified operator will generate keys  |
| ValidatorKeys          | Key(PrivKey,PubKey)                 | Optional field to specify validators keys, if not specified operator will generate keys |
| BesuNodeSpec           | BesuNodeSpec(Image,Resources,etc)   | Optional field to specify common besu node configuration                                |
| GenesisJSON            | GenesisJSON                         | Configuration for genesis block                                                         |
| BootnodesAreValidators | Boolean                             | Specifies whether bootnodes are validators or not, by default set to false              |
| Monitoring             | Boolean                             | Specifies whether to deploy grafana/prometheus or not                                   |
| PrometheusSpec         | PrometheusSpec(Image,Resources,etc) | Prometheus configuration options                                                        |
| GrafanaSpec            | GrafanaSpec(Image,Resources,etc)    | Grafana configuration options                                                           |

2. BesuNode

   Spec Fields

| Field     | Type                                       | Description                                            |
| --------- | ------------------------------------------ | ------------------------------------------------------ |
| Type      | String:["Member", "Bootnode", "Validator"] | Specifies type of besunode                             |
| Replicas  | Integer                                    | Number of replicas of pods of the besunode             |
| Image     | Image                                      | Specifies image repository, tag, and image pull policy |
| Resources | Resources                                  | Specifies CPU & Memory requests & limits               |
| P2P       | PortConfig                                 | Specifies port configurations for P2P                  |
| RPC       | PortConfig                                 | Specifies port configurations for RPC                  |
| WS        | PortConfig                                 | Specifies port configurations for WS                   |
| GraphQl   | PortConfig                                 | Specifies port configurations for GraphQl              |
| Metrics   | PortConfig                                 | Specifies port configurations for Metrics              |

3. Grafana

   Spec Fields

| Field     | Type      | Description                                            |
| --------- | --------- | ------------------------------------------------------ |
| Replicas  | Integer   | Number of replicas of pods of the besunode             |
| Image     | Image     | Specifies image repository, tag, and image pull policy |
| Resources | Resources | Specifies CPU & Memory requests & limits               |
| NodePort  | Integer   | Specifies nodeport                                     |

4. Prometheus

   Spec Fields

| Field     | Type      | Description                                            |
| --------- | --------- | ------------------------------------------------------ |
| Replicas  | Integer   | Number of replicas of pods of the besunode             |
| Image     | Image     | Specifies image repository, tag, and image pull policy |
| Resources | Resources | Specifies CPU & Memory requests & limits               |
| NodePort  | Integer   | Specifies nodeport                                     |

### Useful commands

1. Build docker image corresponding to operator : `operator-sdk build <docker repo>`
2. Push docker image to docker hub repository : `docker push <docker repo>`
3. Make sure to update tag in operator.yaml before deploying any new version
4. If types are changed
   - `operator generate crds; operator generate k8s`
   - Redeploy corresponding CRDs
