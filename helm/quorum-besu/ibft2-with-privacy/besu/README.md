
## Flow of the process:
- Create private/public keys for the bootnodes & update the secrets/bootnode-keys-secret.yaml with the bootnode private keys
- Update the values.yaml file with specific data for the genesis file, number of nodes etc
- Run helm
  - Creates a secrets resource for the validator private keys
  - Creates a configmap with the validator public keys & genesis file.json
  - Spins up the validators up with the private keys specified & associated services for each
  - Spins up the nodes (rpc + ws) and a single service to communicate with the nodes
- Monitoring via prometheus & grafana is also setup up in a separate *monitoring* namespace and exposed via NodePort services (ports 30090, 30030 respectively)
- Credentials for grafana are admin:password. When grafana loads up select the "Besu Dashboard"


## Overview of Setup
![Image ibft](../../../images/ibft-orion.png)

## NOTE:
1. validators1 and 2 serve as bootnodes as well. Adjust according to your needs
2. If you add more validators in past the initial setup, they need to be voted in to be validators i.e they will serve as normal nodes and not validators until they've been voted in.
3. Node1Privacy to Node2Privacy are tied to Orion1 to Orion2 respectively

## Pre chart install - you need to create config that you want to persist
#### 1. Validators private keys
Create private/public keys for the validators using the besu subcommands. The private keys are put into secrets and the public keys go into a configmap to get the bootnode enode address easily
Update the `count` key in the ibftConfigFile.json to alter the number of validators as you would like to provision i.e keys and replicate the deployment & service

```bash
docker run --rm --volume $PWD/ibftSetup/:/opt/besu/data hyperledger/besu:latest operator generate-blockchain-config --config-file=/opt/besu/data/ibftConfigFile.json --to=/opt/besu/data/networkFiles --private-key-file-name=key
sudo chown -R $USER:$USER ./ibftSetup
cp ./ibftSetup/networkFiles/genesis.json ./
```

Update the values.yaml with the keys. The private keys are put into secrets and the public keys go into a configmap that other nodes use to create the enode address

#### 2. Genesis.json
The genesis.json file generated should have been placed at the root directory of this helm chart

#### 3. Orion keys
For more information please refer to the [documentation](https://docs.orion.pegasys.tech/en/stable/Getting-Started/Quickstart/#2-generate-keys) 
Create the keypairs and enter the password when requested. 
```bash
docker run -it --volume $PWD/orionSetup/orion1:/opt/orion/data --entrypoint "/bin/sh" pegasyseng/orion:latest -c 'cd /opt/orion/data && cat orion1.password | /opt/orion/bin/orion --generatekeys nodeKey'
docker run -it --volume $PWD/orionSetup/orion2:/opt/orion/data --entrypoint "/bin/sh" pegasyseng/orion:latest -c 'cd /opt/orion/data && cat orion2.password | /opt/orion/bin/orion --generatekeys  nodeKey' 
sudo chown -R $USER:$USER ./orionSetup
```

#### 4. Orion configuration
Update the orion<n>.conf files to suit requirements 

#### 5. Update any more config in values.yaml if required eg: volume sizes, alter the number of nodes on the network etc
Update the number to nodes to suit, the key is
```bash
node:
  replicaCount: 1
```

#### 6. Run helm and install the chart
```bash
helm install besu ./besu
```

#### 7. In the dashboard, you will see each bootnode deployment & service, nodes & a node service, miner if enabled, secrets(opaque) and a configmap

If using minikube
```bash
minikube dashboard &
```

#### 8. Verify that the nodes are communicating:
```bash
minikube ssh

# once in the terminal
curl -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' <BESU_NODE_SERVICE_HOST>:8545

# which should return:
The result confirms that the node running the JSON-RPC service has two peers:
{
  "jsonrpc" : "2.0",
  "id" : 1,
  "result" : "0x5"
}

```

#### 9. Monitoring
Get the ip that minikube is running on
```bash
minikube ip
```

For example if the ip returned was `192.168.99.100`

*Prometheus:*
In a fresh browser tab open `192.168.99.100:30090` to get to the prometheus dashboard and you can see all the available metrics, as well as the targets that it is collecting metrics for

*Grafana:*
In a fresh browser tab open `192.168.99.100:30030` to get to the grafana dashboard. Credentials are `admin:password` Open the 'Besu Dashboard' to see the status of the nodes on your network. If you do not see the dashboard, click on Dashboards -> Manage and select the dashboard from there


#### 10. Delete
```bash
helm del besu

```
