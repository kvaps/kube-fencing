#!/bin/sh

warn () {
  echo "$@" >&2
}

die () {
  status="$1"
  shift
  warn "$@"
  exit "$status"
}


echo "Starting loop"
unset column
# kubectl get node -o 'custom-columns=NAME:.metadata.name,STATUS:.status.conditions[?(@.type=="Ready")].status'
kubectl get node -o json -w  | 
  jq --raw-output --unbuffered '
  .metadata.name,
  .metadata.creationTimestamp,
  (.status.conditions[] | select(.type=="Ready") | .status),
  (.status.conditions[] | select(.type=="Ready") | .lastHeartbeatTime)' |
while read string; do
  column=$((column+1))
  case "${column}" in
    1)
      name="${string}"
      continue
      ;;
    2)
      status="${string}"
      unset column
      ;;
  esac

  if [ "${status}" = "True" ]; then
      continue
  fi

  #FENCING_AGENT=$(kubectl get pod -l "${FENCING_AGENT_SELECTOR}" | awk '$3 == "Running" {print $1}' | head -n1)

  #kubectl get node -o json "$node" | jq 'del(.status)' > /tmp/node.json
  #fence_command
  #kubectl delete node "$node"
  #kubectl create -f /tmp/node.json


done

