#!/bin/sh
set -e

cd $(dirname $0)

for image in $(find . -maxdepth 2 -name Dockerfile | awk -F/ '{print $2}'); do
  (set -x; docker build --build-arg VERSION=${CI_COMMIT_REF_NAME-latest} -t ${CI_REGISTRY_IMAGE:-kvaps}/kube-$image:${CI_COMMIT_REF_NAME:-latest} $image)
done
