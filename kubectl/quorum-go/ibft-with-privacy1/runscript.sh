#!/bin/bash
if [ -z $1 ] || [ ! -f $1 ]; then
  echo "Please provide a valid script file to execute as the first parameter (i.e. private_contract.js)" >&2
  exit 1
fi

NODE_NUMBER=1
POD=$(kubectl get pod --field-selector=status.phase=Running -o name | grep quorum-node$NODE_NUMBER)
kubectl $NAMESPACE exec -it $POD -c quorum -- /etc/quorum/qdata/contracts/runscript.sh /etc/quorum/qdata/contracts/$1
