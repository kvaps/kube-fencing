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
kubectl get node -o json -w -l fencing=enabled | 
  jq --raw-output --unbuffered '
  .metadata.name,
  (.status.conditions[] | select(.type=="Ready") | .status)' |
while read string; do
  column=$((column+1))
  case "${column}" in
    1)
      name="${string}"
      continue
      ;;
    2)
      ready="${string}"
      unset column
      ;;
  esac
;;

  if [ "${ready}" != "False" ]
  then
      continue
  fi

  #FENCING_AGENT=$(kubectl get pod -l "${FENCING_AGENT_SELECTOR}" | awk '$3 == "Running" {print $1}' | head -n1)

done

