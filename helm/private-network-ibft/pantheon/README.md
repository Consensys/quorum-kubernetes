
## Flow of the process:
- Create private/public keys for the bootnodes & update the secrets/bootnode-keys-secret.yaml with the bootnode private keys
- Update the values.yaml file with specific data for the genesis file, number of nodes etc
- Run helm
  - Creates a secrets resource for the validator private keys
  - Creates a configmap with the validator public keys & genesis file.json
  - Spins up the validators up with the private keys specified & associated services for each
  - Spins up the nodes (rpc + ws) and a single service to communicate with the nodes

## NOTE:
1. validators1 and 2 serve as bootnodes as well. Adjust according to your needs
2. If you add more validators in past the initial setup, they need to be voted in to be validators i.e they will serve as normal nodes and not validators until they've been voted in.

## Pre chart install - you need to create config that you want to persist
#### 1. Validators private keys
Create private/public keys for the validators using the pantheon subcommands. The private keys are put into secrets and the public keys go into a configmap to get the bootnode enode address easily
Repeat this process for as many validators as you would like to provision i.e keys and replicate the deployment & service

```bash
docker run --rm --volume $PWD/ibftSetup/:/opt/pantheon/data pegasyseng/pantheon:develop operator generate-blockchain-config --config-file=/opt/pantheon/data/ibftConfigFile.json --to=/opt/pantheon/data/networkFiles --private-key-file-name=key
sudo chown -R $USER:$USER ./ibftSetup
mv ./ibftSetup/networkFiles/genesis.json ./ibftSetup/
```

Update the values.yaml with the keys. The private keys are put into secrets and the public keys go into a configmap that other nodes use to create the enode address

#### 2. Genesis.json
Copy the genesis.json file generated and place it at the root directory of this helm chart

#### 3. Update any more config in values.yaml if required eg: volume sizes, alter the number of nodes on the network etc
Update the number to nodes to suit, the key is
```bash
node:
  replicaCount: 1
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
  "result" : "0x5"
}

```

#### 7. Delete
```bash
helm del --purge pantheon

```
