#!/bin/sh

debug() {
  [ -n "$DEBUG" ] && echo $(date '+%b %d %X') "debug: $@"
}

log() {
  echo $(date '+%b %d %X') "info:  $@"
}

warn() {
  echo $(date '+%b %d %X') "warn:  $@" >&2
}

die() {
  status="$1"
  shift
  echo $(date '+%b %d %X') "fatal: $@" >&2
  exit "$status"
}

main() {
  debug "FENCING_AGENT_SELECTOR=$FENCING_AGENT_SELECTOR"
  debug "FENCING_NODE_SELECTOR=$FENCING_NODE_SELECTOR"
  log "Starting loop"
  kubectl get node -w -l "$FENCING_NODE_SELECTOR" | 
  while read line; do
    while read NAME STATUS ROLES AGE VERSION; do
      debug "$NAME - $STATUS"
      if [ "$STATUS" = "Ready" ]; then
        continue
      fi
      REASON=$(kubectl get node "$NAME" -o 'custom-columns=STATUS:.status.conditions[?(@.type=="Ready")].reason' | tail -n1)
      if [ "$REASON" = "NodeStatusUnknown" ]; then
        log "$NAME - $REASON"
      fi
    done
  done
}

  #FENCING_AGENT=$(kubectl get pod -l "${FENCING_AGENT_SELECTOR}" | awk '$3 == "Running" {print $1}' | head -n1)

  #kubectl get node -o json "$node" | jq 'del(.status)' > /tmp/node.json
  #fence_command
  #kubectl delete node "$node"
  #kubectl create -f /tmp/node.json
