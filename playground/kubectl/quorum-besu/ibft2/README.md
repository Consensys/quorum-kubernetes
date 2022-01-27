
# Example setup using local permissioning and privacy

## Flow of the process:
- Create private/public keys for the validators & update the secrets/validator-keys-secret.yaml with the validator private keys
- Update the local permissions file: configmap/besu-node-permissions-configmap.yml (becuase this is a permissioned setup you must add enode details for every node here)
- Update the configmap/configmap.yml with the public keys & genesis file
- Update the number of nodes you would like in deployments/node-deployment.yaml
- Run kubectl
- Monitoring via prometheus & grafana is also setup up in a separate *monitoring* namespace and exposed via NodePort services (ports 30090, 30030 respectively)
- Credentials for grafana are admin:password. When grafana loads up select the "besu Dashboard"

## Overview of Setup
![Image ibft](../../../../static/ibft-tessera.png)

## NOTE:
1. validators1 and 2 serve as bootnodes as well. Adjust according to your needs
2. If you add more validators in past the initial setup, they need to be voted in to be validators i.e they will serve as normal nodes and not validators until they've been voted in.

#### 1. Boot nodes private keys 
**Note:** Please note that steps 1-3 have already been run and the k8s objects already have keys. You are welcome to run steps 1-3 yourself though and create fresh keys

Create private/public keys for the validators using the besu subcommands. The private keys are put into secrets and the public keys go into a configmap to get the bootnode enode address easily
Repeat this process for as many validators as you would like to provision i.e keys and replicate the deployment & service

Essentially you are following the [genesis tutorial](https://besu.hyperledger.org/en/latest/Tutorials/Private-Network/Create-IBFT-Network/) here. Create a folder called `ibftSetup` and copy the `ibftConfigFile.json`
from the tutorial in there to run the key generation process. 

```bash
mkdir -p $PWD/ibftSetup
# copy the `ibftConfigFile.json` file from the tutorial here and then proceed
docker run --rm --volume $PWD/ibftSetup/:/opt/besu/data hyperledger/besu:latest operator generate-blockchain-config --config-file=/opt/besu/data/ibftConfigFile.json --to=/opt/besu/data/networkFiles --private-key-file-name=key
sudo chown -R $USER:$USER ./ibftSetup
mv ./ibftSetup/networkFiles/genesis.json ./ibftSetup/
```

Update the secrets/validator-key-secret.yaml with the private keys. The private keys are put into secrets and the public keys go into a configmap that other nodes use to create the enode address
Update the configmap/configmap.yaml with the public keys
**Note:** Please remove the '0x' prefix of the public keys

#### 2. Genesis.json
Copy the genesis.json file and copy its contents into the configmap/configmap as shown

#### 3. Update any more config if required
eg: To alter the number of nodes on the network, alter the `replicas: 2` in the deployments/node-deployments.yaml to suit

#### 4. Deploy:
```bash

./deploy.sh

```


#### 5. In the dashboard, you will see each bootnode deployment & service, nodes & a node service, miner if enabled, secrets(opaque) and a configmap

If using minikube
```bash
minikube dashboard &
```

#### 6. Verify that the nodes are communicating:
```bash
minikube ssh

# once in the terminal
curl -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' <besu_NODE_SERVICE_HOST>:8545

# which should return:
The result confirms that the node running the JSON-RPC service has two peers:
{
  "jsonrpc" : "2.0",
  "id" : 1,
  "result" : "0x5"
}

```


#### 7. Monitoring
Get the ip that minikube is running on
```bash
minikube ip
```

For example if the ip returned was `192.168.99.100`

*Prometheus:*
In a fresh browser tab open `192.168.99.100:30090` to get to the prometheus dashboard and you can see all the available metrics, as well as the targets that it is collecting metrics for

*Grafana:*
In a fresh browser tab open `192.168.99.100:30030` to get to the grafana dashboard. Credentials are `admin:password` Open the 'besu Dashboard' to see the status of the nodes on your network. If you do not see the dashboard, click on Dashboards -> Manage and select the dashboard from there


#### 8. Delete
```
./remove.sh
```
