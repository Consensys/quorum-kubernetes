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
POD=$(kubectl get pod --field-selector=status.phase=Running -o name | grep quorum-node$NODE_NUMBER)
kubectl $NAMESPACE exec -it $POD -c quorum -- /geth-helpers/geth-attach.sh