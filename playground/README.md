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
Once you have installed the tools above, `cd` into the client (besu/goquorum) and into the consensus algorithm of choice and then run `./deploy.sh`

To remove the k8s objects, run `./remove.sh`