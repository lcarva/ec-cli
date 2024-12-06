#!/usr/bin/env bash
# Copyright The Enterprise Contract Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0

# This script is meant to take an existing snapshot reference which includes just
# the EC CLI image and expand that into a new snapshot which includes both the
# EC CLI image and the EC Tekton bundle image.

set -o errexit
set -o nounset
set -o pipefail

SNAPSHOT_NAME=$1
NEW_SNAPSHOT_PATH=$2

function debug() {
    echo "[DEBUG] $1" >&2
}

debug "Fetching ${SNAPSHOT_NAME} snapshot"
SNAPSHOT_SPEC="$(oc get snapshot ${SNAPSHOT_NAME} -o json | jq '.spec')"
debug "${SNAPSHOT_SPEC}"

debug "Verifying snapshot contains a single component"
echo "${SNAPSHOT_SPEC}" | jq -e '.components | length == 1' > /dev/null

CLI_IMAGE_REF="$(echo "${SNAPSHOT_SPEC}" | jq -r '.components[0].containerImage')"
debug "CLI image ref: ${CLI_IMAGE_REF}"

BUNDLE_IMAGE_REF="$(
    cosign download attestation "${CLI_IMAGE_REF}" | jq -r '.payload | @base64d | fromjson |
        .predicate.buildConfig.tasks[] | select(.name == "build-tekton-bundle") |
        .results[] | select(.name == "IMAGE_REF") | .value'
)"

debug "Bundle image ref: ${BUNDLE_IMAGE_REF}"

debug "Creating new snapshot spec"

echo "${SNAPSHOT_SPEC}" | jq  --arg bundle "${BUNDLE_IMAGE_REF}" \
    '.components[0].source as $source | .components += [{
        "name": "tekton-bundle", "containerImage": $bundle, "source": $source
    }]' | tee "${NEW_SNAPSHOT_PATH}"
