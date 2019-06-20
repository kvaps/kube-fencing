#!/bin/sh

log() {
  echo $(date '+%b %d %X') "info:  $@"
}

warn() {
  echo $(date '+%b %d %X') "warn:  $@" >&2
}

die() {
  status="$1"
  shift
  echo $(date '+%b %d %X') "error: $@" >&2
  exit "$status"
}

debug() {
  [ -n "$DEBUG" ] && echo $(date '+%b %d %X') "debug: $@" >&2
}

# Run external command
run() {
  set -e
  ( [ -n "$DEBUG" ] && set -x; "$@" )
  set +e
}

# Fencing node via fencing agent pod
fence() {
  log "Fencing node $1"
  FENCING_AGENT_POD=$(run kubectl get pod -l "${FENCING_AGENT_SELECTOR}" | awk '$3 == "Running" {print $1}' | head -n1)
  if [ -z "$FENCING_AGENT_POD" ]; then
    die 1 "Can not find running fencing agent pod"
  fi
  debug "FENCING_AGENT_POD: $FENCING_AGENT_POD"
  run kubectl exec "$FENCING_AGENT_POD" -- $FENCING_SCRIPT "$1"
  if [ $? -eq 126 ]; then
    die 2 "Can not find fencing script"
  fi
}

# Flush node in kubernetes
flush() {
  set -e
  case "$FLUSHING_MODE" in
    none)
      true
    ;;
    delete)
      log "Deleting node $1"
      run kubectl delete node "$1"
    ;;
    recreate)
      log "Deleting pods from $1"
      run kubectl get pod --field-selector "spec.nodeName=$1" --all-namespaces -o custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name |
        awk '{if (!a[$1]++){printf "\n" $1 " " $2} else {printf " " $2}} END {print ""}' | tail -n+3 |
        while read line; do
          NAMESPACE="${line%% *}"
          PODS="${line#* }"
          run kubectl delete pod -n "$NAMESPACE" $PODS --grace-period=0 --force --wait=false 2>/dev/null
        done
      log "Recreating node $1"
      run kubectl get node -o json "$1" | jq 'del(.status)' > "/tmp/node-${1}.yaml"
      run kubectl delete node "$1"
      run kubectl create -f "/tmp/node-${1}.yaml"
      rm -f "/tmp/node-${1}.yaml"
    ;;
  esac
  set +e
}

wait_back_online() {
  local START=$(date +%s)
  run kubectl wait node --for=condition=Ready --timeout=${2}s "$1"
  local EC=$?
  local END=$(date +%s)
  echo "$((END-START))"
  return "$EC"
}

main() {
  log "Loading parameters"
  FENCING_NODE_SELECTOR=${FENCING_NODE_SELECTOR:-fencing=enabled}
  FENCING_AGENT_SELECTOR=${FENCING_AGENT_SELECTOR:-app=fencing-agents}
  FENCING_SCRIPT=${FENCING_SCRIPT:-/scripts/fence.sh}
  FLUSHING_MODE=${FLUSHING_MODE:-delete}
  TIMEOUT=${TIMEOUT:-0}
  log "FENCING_NODE_SELECTOR: $FENCING_NODE_SELECTOR"
  log "FENCING_AGENT_SELECTOR: $FENCING_AGENT_SELECTOR"
  log "FENCING_SCRIPT: $FENCING_SCRIPT"
  log "FLUSHING_MODE: $FLUSHING_MODE"
  log "TIMEOUT: $TIMEOUT"

  log "Starting loop"
  run kubectl get node -w -l "$FENCING_NODE_SELECTOR" | 
  while read line; do
    while read NAME STATUS ROLES AGE VERSION; do
      debug "$NAME - ${STATUS%%,*}"
      if [ "${STATUS%%,*}" = "Ready" ]; then
        continue
      fi
      REASON=$(kubectl get node "$NAME" -o 'custom-columns=STATUS:.status.conditions[?(@.type=="Ready")].reason' | tail -n1)
      if [ "$REASON" != "NodeStatusUnknown" ]; then
        PREVIOUS_NAME=
        continue
      fi
      if [ "$NAME" = "$PREVIOUS_NAME" ]; then
        continue
      fi
      PREVIOUS_NAME="$NAME"
      log "$NAME - $REASON"
      if [ "$FLUSHING_MODE" = "info" ]; then
        continue
      fi

      log "Waiting $TIMEOUT seconds until node come back"
      if [ "$TIMEOUT" -gt 0 ] 2>/dev/null; then
        WAIT_TIME=$(wait_back_online "$NAME" "$TIMEOUT")
        if [ $? -eq 0 ]; then
          log "Node $NAME come back online after $WAIT_TIME seconds"
          log "Fencing aborted"
          break
        else
          log "Node $NAME has not come back in $TIMEOUT seconds timeout period"
        fi
      fi

      fence "$NAME"
      if [ $? -eq 0 ]; then
        log "Fencing success"
        flush "$NAME"
      else
        warn "Fencing failed"
      fi

    done
  done
}

main "$@"
