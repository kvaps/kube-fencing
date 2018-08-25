#!/bin/sh
log() {
  echo $(date '+%b %d %X') "info:  $@"
}

die() {
  status="$1"
  shift
  echo $(date '+%b %d %X') "error: $@" >&2
  exit "$status"
}

# Run external command
run() {
  set -e
  ( [ -n "$DEBUG" ] && set -x; "$@" )
  set +e
}

enable_fencing() {
  log "Enabling fencing"
  run kubectl label node --overwrite "${NODE_NAME}" "${FENCING_LABEL}=enabled"
}

disable_fencing() {
  log "Disabling fencing"
  run kubectl label node --overwrite "${NODE_NAME}" "${FENCING_LABEL}=disable"
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
