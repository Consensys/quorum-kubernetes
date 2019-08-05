
## Flow of the process:
- Create private/public keys for the bootnodes & update the values.yaml with the bootnode keys
- Update the values.yaml file with specific data for the genesis file
- Run helm
  - Creates a secrets resource for the bootnode private keys
  - Creates a configmap with the bootnode public keys & genesis file.json
  - Spins up the bootnodes with the private keys specified & associated services for each
  - Spins up the nodes (rpc + ws) and a single service to communicate with the nodes
  - Spins up a miner node if enabled
- Monitoring via prometheus & grafana is also setup up in a separate *monitoring* namespace and exposed via NodePort services (ports 30090, 30030 respectively)
- Credentials for grafana are admin:admin. When grafana loads up select the "Pantheon Dashboard"


## Pre chart install - you need to create config that you want to persist

#### 1. Boot nodes private keys
Create private/public keys for the bootnodes using the pantheon subcommands. The private keys are put into secrets and the public keys go into a configmap to get the bootnode enode address easily
Repeat this process for as many bootnodes as you would like to provision i.e keys and replicate the deployment & service

```bash
mkdir -p ./generated_config/bootnode1 ./generated_config/bootnode2
docker run --rm --volume $PWD/generated_config/bootnode1:/opt/pantheon/data pegasyseng/pantheon:develop --data-path /opt/pantheon/data public-key export --to /opt/pantheon/data/key.pub
docker run --rm --volume $PWD/generated_config/bootnode2:/opt/pantheon/data pegasyseng/pantheon:develop --data-path /opt/pantheon/data public-key export --to /opt/pantheon/data/key.pub
```

Update the values.yaml with the keys. The private keys are put into secrets and the public keys go into a configmap that other nodes use to create the enode address

#### 2. Genesis.json
Create the genesis.json file and place it at the root directory of this helm chart

#### 3. Update any more config in values.yaml if required eg: alter the number of nodes on the network.
Update the number to nodes to suit, the key is
```bash
node:
  replicaCount: 2
```

#### 4. Run helm
```bash
helm install --namespace NAMESPACE --name pantheon ./pantheon
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
curl -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' <PANTHEON_NODE_SERVICE_HOST>:8545

# which should return:
The result confirms that the node running the JSON-RPC service has two peers:
{
  "jsonrpc" : "2.0",
  "id" : 1,
  "result" : "0x3"
}

```

#### 7. Delete
```bash
helm del --purge pantheon

```
