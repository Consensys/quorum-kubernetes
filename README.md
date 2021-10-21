

# Quorum-Kubernetes (k8s)

The following repo has example reference implementations of private networks using k8s. These examples are aimed at developers and ops people to get them familiar with how to run a private ethereum network in k8s and understand the concepts involved.

You will need the following tools to proceed:

- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/) This is the local equivalent of a K8S cluster (refer to the [playground](./playground) for manifests to deploy)
- [Helm](https://helm.sh/docs/)
- [Helmfile](https://github.com/roboll/helmfile)
- [Helm Diff plugin](https://github.com/databus23/helm-diff)

Verify kubectl is connected with
```bash
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.1", GitCommit:"4485c6f18cee9a5d3c3b4e523bd27972b1b53892", GitTreeState:"clean", BuildDate:"2019-07-18T09:18:22Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.0", GitCommit:"e8462b5b5dc2584fdcd18e6bcfe9f1e4d970a529", GitTreeState:"clean", BuildDate:"2019-06-19T16:32:14Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
```

Install helm & helm-diff:
Please note that the documentation and steps listed use *helm3*. The API has been updated so please take that into account if using an older version
```bash
$ helm plugin install https://github.com/databus23/helm-diff --version master
```

The repo provides examples using multiple tools such as kubectl, helm, helmfile etc. Please select the one that meets your deployment requirements.

The current repo layout is:

```bash
  ├── docker
  │   └── istanbul-tools            # helper docker images used for various tasks
  ├── ingress                       # ingress rules, hidden here for brevity
  │   ├── ...                       
  ├── azure                         # azure specific artifacts
  │   ├── arm                       # azure ARM templates to deploy resources ie cluster, keyvault, identity etc
  │   └── scripts                   # azure scripts to install CSI drivers on the AKS cluster and the like
  ├── playground                    # playground for users to get familiar with concepts and how to run and tweak things - START HERE 
  │   └── kubectl
  │       ├── quorum-besu           # use Hyperledger Besu as the block chain client
  │       │   ├── clique
  │       │   │   ├── ...           # templates, config etc hidden here for brevity
  │       │   ├── ethash
  │       │   │   ├── ...
  │       │   └── ibft2
  │       │       └── ...
  │       └── quorum-go             # use GoQuorum as the block chain client
  │           └── ibft
  │               └── ...
  ├── dev                       
  │   └── helm  
  │       ├── charts            
  │       │   ├── ...               # dev helm charts, hidden here for brevity
  │       └── values            
  │           ├── ...               # values.yml overrides for various node types
  ├── prod                      
  │   └── helm  
  │       ├── charts            
  │       │   ├── ...              # prod helm charts - these will use cloud native services where possible eg IAM for identity, keyvault for secrets etc
  │       └── values            
  │           ├── ...              # values.yml overrides for various node types
  └── static                       # images and other static assets
  


```

We recommend starting with the `playground` folder and working through the example setups there and then moving to the next `dev` stage.

The `dev` and `prod` folders are pretty identical in terms of what gets deployed, but differ in that the prod folder natively uses best practices to manage identity (Managed Identities in Azure and IAM in AWS) and vaults (Keyvault in Azure and KMS in AWS) along with CSI drivers

## Concepts:

#### Providers
If you are deploying to Azure, please refer to the ARM templates and deployment [documentation](./azure/README.md)

If you are deploying locally you need a Kubernetes cluster like [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/)

#### Namespaces:
Currently we do not deploy anything in the 'default' namespace. Anything related to Besu gets spun up in a 'besu' namespace, and 'quorum' for GoQuorum; with the monitoring pieces get spun up in a 'monitoring' namespace.
Namespaces are part of the setup and do not need to be created via kubectl prior to deploying. To change the namespaces:
- In Kubectl, you need to edit every file in the deployment
- In Helm, edit the namespace value in the values.yaml 

It is recommended you follow this approach as well in your production setups and where possible use Service Accounts to secure deployments & statefulsets. We make use of these extensively.

#### Network Topology and High Availability requirements:
Ensure that if you are using a cloud provider you have enough spread across AZ's to minimize risks - refer to our [HA](https://besu.hyperledger.org/en/latest/HowTo/Configure/Configure-HA/High-Availability/) and [Load Balancing] (https://besu.hyperledger.org/en/latest/HowTo/Configure/Configure-HA/Sample-Configuration/) documentation

When deploying a private network, eg: IBFT you need to ensure that the bootnodes are accessible to all nodes on the network. Although the minimum number needed is 1, we recommend you use more than 1 spread across AZ's. In addition we also recommend you spread validators across AZ's and have a sufficient number available in the event of an AZ going down.

You need to ensure that the genesis file is accessible to all nodes joining the network.

Hyperledger Besu supports [NAT mechanisms](https://besu.hyperledger.org/en/stable/Reference/CLI/CLI-Syntax/#nat-method) and the default is set to automatically handle NAT environments. If you experience issues with NAT and logs have messages that have the NATService throwing exceptions connecting to external IPs, please add this option in your Besu deployments `--nat-method = NONE`

#### Data Volumes:
Ensure that you provide enough capacity for data storage for all nodes that are going to be on the cluster. Select the appropriate [type](https://kubernetes.io/docs/concepts/storage/volumes/) of persitent volume based on your cloud provider.

#### Nodes:
Consider the use of statefulsets instead of deployments for nodes. The term 'node' refers to bootnode, validator and network nodes.

Configuration of nodes can be done either via a single item inside a config map, as Environment Variables or as command line options. Please refer to the [Configuration](https://besu.hyperledger.org/en/latest/HowTo/Configure/Using-Configuration-File/) section of our documentation

#### RBAC:
We encourage the use of RBAC's for access to the private key of each node, ie. only a specific pod/statefulset is allowed to access a specific secret. If you need to specify a Kube config file to each pod please use the `KUBE_CONFIG_PATH` variable

#### Monitoring
As always please ensure you have sufficient monitoring and alerting setup.

Besu & GoQuorum publish metrics to [Prometheus](https://prometheus.io/) and metrics can be configured using the [kubernetes scraper config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config).

Besu & GoQuorum also have a custom Grafana [dashboards](https://grafana.com/orgs/pegasyseng) to make monitoring of the nodes easier.

For ease of use, the kubectl & helm examples included have both installed and included as part of the setup. Please configure the kubernetes scraper and grafana security to suit your requirements, grafana supports multiple options that can be configured using env vars

#### Ingress Controllers:

If you require the use of ingress controllers for the RPC calls or the monitoring dashboards, we have provided [examples](./ingress) with rules that are configured to do so.

Please use these as a reference and develop solutions to match your network topology and requirements.

#### Logging
Node logs can be [configured](https://besu.hyperledger.org/en/latest/HowTo/Troubleshoot/Logging/#advanced-custom-logging) to suit your environment. For example, if you would like to log to file and then have parsed via logstash into an ELK cluster, please follow out documentation.



## New nodes joining the network:
The general rule is that any new nodes joining the network need to have the following accessible:
- genesis.json of the network
- Bootnodes need to be accessible on the network (if using bootnodes, otherwise static-nodes.json)
- Bootnodes enode's (public key and IP) should be passed in at boot
- If you’re using permissioning on your network, specifically authorise the new nodes

If the initial setup was on Kubernetes, you have the following scenarios:

#### 1. New node also being provisioned on the K8S cluster:
In this case anything that applies to how current nodes are provisioned should be applicable and the only thing that need be done is increase the number of replicas

#### 2. New node being provisioned elsewhere
Ensure that the host being provisioned can find and connect to the bootnode's. You may need to use `traceroute`, `telnet` or the like to ensure you have connectivity. Once connectivity has been verified, you need to pass the enode of the bootnodes and the genesis file to the node. This can be done in many ways, for example query the k8s cluster via APIs prior to joining if your environment allows for that. Alternatively put this data somewhere accessible to new nodes that may join in future as well, and pass the values in at runtime.

Ensure that the host being provisioned can also connect to the other nodes that you have on the k8s cluster, otherwise it will be unable to connect to any peers (bar the bootnodes). The most reliable way to do this is via a VPN so it has access to the bootnodes as well as any nodes on the k8s cluster. You can alternatively use ingresses on the nodes (ideally more than just bootnodes) you wish to expose, where TCP & UDP on port 30303 need to be open for discovery.

Additionally if you’re using permissioning on your network you will also have to specifically authorise the new nodes


## Production Network Guidelines:
| ⚠️ **Note**: After you have familiarised yourself with the examples in this repo, it is recommended that you design your network based on your needs, taking the following guidelines into account |
| --- |


#### Pod Resources:
The templates in this repository have been set to run locally on Minikube to get the user familiar with the setup. Hence the resources are set low, when designing your setup to run in `staging` or `production` environments, please ensure you **grant at least 4GB of memory to Besu pods and 2GB of memory to Tessera pods.** Also ensure you **select the appropriate storage class and size for your nodes.**

Ensure that if you are using a cloud provider you have enough spread across AZ's to minimize risks - refer to our [HA](https://besu.hyperledger.org/en/latest/HowTo/Configure/Configure-HA/High-Availability/) and [Load Balancing] (https://besu.hyperledger.org/en/latest/HowTo/Configure/Configure-HA/Sample-Configuration/) documentation

When deploying a private network, eg: IBFT you need to ensure that the bootnodes are accessible to all nodes on the network. Although the minimum number needed is 1, we recommend you use more than 1 spread across AZ's. In addition we also recommend you spread validators across AZ's and have a sufficient number available in the event of an AZ going down.

You need to ensure that the genesis file is accessible to all nodes joining the network.


