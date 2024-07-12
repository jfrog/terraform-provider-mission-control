#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.84.15}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}" > /dev/stderr

# docker cp doesn't support copying files between containers so copy to local disk first
CONTAINER_ID_1=$(docker ps -q --filter "ancestor=releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}" --filter publish=8082)
CONTAINER_ID_2=$(docker ps -q --filter "ancestor=releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}" --filter publish=9082)

echo "Fetching root certificates"
docker cp "${CONTAINER_ID_1}":/opt/jfrog/artifactory/var/etc/access/keys/root.crt "${SCRIPT_DIR}/artifactory-1.crt" \
  && chmod go+rw "${SCRIPT_DIR}"/artifactory-1.crt
docker cp "${CONTAINER_ID_2}":/opt/jfrog/artifactory/var/etc/access/keys/root.crt "${SCRIPT_DIR}/artifactory-2.crt" \
  && chmod go+rw "${SCRIPT_DIR}"/artifactory-2.crt

echo "Uploading root certificates"
docker cp "${SCRIPT_DIR}/artifactory-1.crt" "${CONTAINER_ID_2}:/opt/jfrog/artifactory/var/etc/access/keys/trusted/artifactory-1.crt"
docker cp "${SCRIPT_DIR}/artifactory-2.crt" "${CONTAINER_ID_1}:/opt/jfrog/artifactory/var/etc/access/keys/trusted/artifactory-2.crt"

echo "Circle-of-Trust is setup between artifactory-1 and artifactory-2 instances"