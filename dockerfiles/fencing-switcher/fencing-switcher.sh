#!/bin/sh
set -e

log() {
  echo $(date '+%b %d %X') "info:  $@"
}

die() {
  status="$1"
  shift
  echo $(date '+%b %d %X') "error: $@" >&2
  exit "$status"
}

set_node_label() {
  log "Setting $1 $2=$3"
  CA_CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
  TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
  DATA="[{\"op\": \"add\", \"path\": \"/metadata/labels/$2\", \"value\": \"$3\"}]"
  URL="https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/nodes/$1"
  curl -o /dev/null -sS -m5 --cacert "$CA_CERT" -H "Authorization: Bearer $TOKEN" --request PATCH --data "$DATA" -H "Content-Type:application/json-patch+json" "$URL"
}

enable_fencing() {
  log "Enabling fencing"
  set_node_label "${NODE_NAME}" "$FENCING_LABEL" enabled
}

disable_fencing() {
  log "Disabling fencing"
  set_node_label "${NODE_NAME}" "$FENCING_LABEL" disabled
}

main() {
  case "" in
    "$NODE_NAME" )
      die 1 "Variable: NODE_NAME can't be empty"
    ;;
    "$FENCING_LABEL" )
      die 1 "Variable: FENCING_LABEL can't be empty"
    ;;
  esac

  enable_fencing
  trap 'disable_fencing' INT TERM KILL
  
  log "Sleeping"
  tail -f /dev/null & wait
}

main "$@"
