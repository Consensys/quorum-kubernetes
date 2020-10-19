#!/bin/bash
NUMBER_OF_NODES=4
NODE_NUMBER=$1
case "$NODE_NUMBER" in ("" | *[!0-9]*)
  echo 'Please provide the number of the node to attach to (i.e. ./attach.sh 2)' >&2
  exit 1
esac

if [ "$NODE_NUMBER" -lt 1 ] || [ "$NODE_NUMBER" -gt $NUMBER_OF_NODES ]; then
  echo "$NODE_NUMBER is not a valid node number. Must be between 1 and $NUMBER_OF_NODES." >&2
  exit 1
fi

minikube ip > /dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ];
then
  IP_ADDRESS=localhost
else
  IP_ADDRESS=$(minikube ip)
fi

QUORUM_PORT=$(kubectl get service quorum-node$1 -o=jsonpath='{range.spec.ports[?(@.name=="rpc-listener")]}{.nodePort}')

TESSERA_PORT=$(kubectl get service quorum-node$1 -o=jsonpath='{range.spec.ports[?(@.name=="tm-tessera-third-part")]}{.nodePort}')

echo quorum rpc: http://$IP_ADDRESS:$QUORUM_PORT
echo tessera 3rd party: http://$IP_ADDRESS:$TESSERA_PORT
