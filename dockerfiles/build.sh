#!/bin/sh

cd $(dirname $0)

for image in $(find . -maxdepth 2 -name Dockerfile | awk -F/ '{print $2}'); do
  export DOCKER_BUILDKIT=1
  docker build --network=host --build-arg VERSION=${CI_COMMIT_REF_NAME-latest} -t ${CI_REGISTRY_IMAGE:-kvaps}/kube-$image:${CI_COMMIT_REF_NAME:-latest} --progress plain $image
done
