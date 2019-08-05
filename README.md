

# Pantheon-k8s

The following repo has example reference implementations of private networks using k8s. This is intended to get developers and ops people familiar with how to run a private ethereum network in k8s and understand the concepts involved.

It provides examples using multiple tools such as kubectl, helm, helmfile etc. Please select the one that meets your deployment requirements.

## Local Development:
The reference examples in this repo can be used locally, to get familiar with the deployment setup. You will require:
- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/) This is the local equivalent of a K8S cluster
- [Helm](https://helm.sh/docs/)
- [Helmfile](https://github.com/roboll/helmfile)
- [Helm Diff plugin](https://github.com/databus23/helm-diff)


Verify kubectl and minikube are working with
```bash
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.1", GitCommit:"4485c6f18cee9a5d3c3b4e523bd27972b1b53892", GitTreeState:"clean", BuildDate:"2019-07-18T09:18:22Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"15", GitVersion:"v1.15.0", GitCommit:"e8462b5b5dc2584fdcd18e6bcfe9f1e4d970a529", GitTreeState:"clean", BuildDate:"2019-06-19T16:32:14Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}
```

Install helm with:
```bash
$ helm init
$ helm plugin install https://github.com/databus23/helm-diff --version master
```

Pick the deployment tool that suits your environment and then change directory and follow the Readme.md files there.



## Production Network Guidelines:
| ⚠️ **Note**: After you have familiarised yourself with the examples in this repo, it is recommended that you design your network based on your needs, taking the following guidelines into account |
| --- |

#### Network Topology and High Availability requirements:
Ensure that if you are using a cloud provider you have enough spread across AZ's to minimize risks - refer to our [HA](https://docs.pantheon.pegasys.tech/en/stable/Deploying-Pantheon/High-Availability/) documentation

When deploying a private network, eg: IBFT you need to ensure that the bootnodes are accessible to all nodes on the network. Although the minimum number needed is 1, we recommend you use more than 1 spread across AZ's. In addition we also recommend you spread validators across AZ's and have a sufficient number available in the event of an AZ going down.

You need to ensure that the genesis file is accessible to all nodes joining the network.

#### Data Volumes:
Ensure that you provide enough capacity for data storage for all nodes that are going to be on the cluster. Select the appropriate [type](https://kubernetes.io/docs/concepts/storage/volumes/) of persitent volume based on your cloud provider.

#### Nodes:
Consider the use of statefulsets instead of depolyments for nodes. The term 'node' refers to bootnode, validator and network nodes.

Configuration of nodes can be done either via a single item inside a config map, as Environment Variables or as command line options. Please refer to the [Configuration](https://docs.pantheon.pegasys.tech/en/stable/Configuring-Pantheon/Using-Configuration-File/) section of our documentation


#### Monitoring
As always please ensure you have sufficient monitoring and alerting setup.

Pantheon publishes metrics to [Prometheus](https://prometheus.io/) and metrics can be configured using the [kubernetes scraper config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config).

Pantheon also has a custom Grafana [dashboard](https://grafana.com/grafana/dashboards/10273) to make monitoring of the nodes easier.

For ease of use, the kubectl & helm examples included have both installed and included as part of the setup. Please configure the kubernetes scraper and grafana security to suit your requirements.

#### Logging
Pantheon's logs can be [configured](https://docs.pantheon.pegasys.tech/en/stable/Monitoring/Logging/#advanced-custom-logging) to suit your environment. For example, if you would like to log to file and then have parsed via logstash into an ELK cluster, please follow out documentation.


