#!/bin/bash

set -e

GITHUB_SHA=${GITHUB_SHA:-"`git rev-parse HEAD`"}
IMAGE_ORG="docker.io/adagioworkflow"
TAG="`echo $GITHUB_SHA | cut -c -8`"

echo "Tagging Images with SHA"

docker tag "${IMAGE_ORG}/adagio:latest" "${IMAGE_ORG}/adagio:${TAG}"
docker tag "${IMAGE_ORG}/ui:latest" "${IMAGE_ORG}/ui:${TAG}"

echo "Pushing Images"

docker push "${IMAGE_ORG}/adagio:latest"
docker push "${IMAGE_ORG}/adagio:${TAG}"
docker push "${IMAGE_ORG}/ui:latest"
docker push "${IMAGE_ORG}/ui:${TAG}"
