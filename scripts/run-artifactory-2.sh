#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
source "${SCRIPT_DIR}/get-access-key.sh"
source "${SCRIPT_DIR}/wait-for-rt.sh"

export ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION:-7.84.15}
echo "ARTIFACTORY_VERSION=${ARTIFACTORY_VERSION}" > /dev/stderr

set -euf

sudo rm -rf ${SCRIPT_DIR}/artifactory-2/

mkdir -p ${SCRIPT_DIR}/artifactory-2/extra_conf
mkdir -p ${SCRIPT_DIR}/artifactory-2/var/etc/access

cp ${SCRIPT_DIR}/artifactory-2.lic ${SCRIPT_DIR}/artifactory-2/extra_conf
cp ${SCRIPT_DIR}/system.yaml ${SCRIPT_DIR}/artifactory-2/var/etc/
cp ${SCRIPT_DIR}/access.config.patch.yml ${SCRIPT_DIR}/artifactory-2/var/etc/access

if [[ -z "${ARTIFACTORY_JOIN_KEY}" ]]; then
  yq -i '.shared += {"security": {"joinKey": "$ARTIFACTORY_JOIN_KEY"}}' ${SCRIPT_DIR}/artifactory-2/var/etc/system.yaml
fi

docker run -i --name artifactory-2 -d --rm \
  -e JF_FRONTEND_FEATURETOGGLER_ACCESSINTEGRATION=true \
  -v ${SCRIPT_DIR}/artifactory-2/extra_conf:/artifactory_extra_conf \
  -v ${SCRIPT_DIR}/artifactory-2/var:/var/opt/jfrog/artifactory \
  -p 9081:8081 -p 9082:8082 \
  releases-docker.jfrog.io/jfrog/artifactory-pro:${ARTIFACTORY_VERSION}

export ARTIFACTORY_URL_2=http://localhost:9081
export ARTIFACTORY_UI_URL_2=http://localhost:9082

# Wait for Artifactory to start
waitForArtifactory "${ARTIFACTORY_URL_2}" "${ARTIFACTORY_UI_URL_2}"
