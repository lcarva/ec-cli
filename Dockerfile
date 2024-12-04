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

## Build

FROM docker.io/library/golang:1.22.7 AS build

ARG TARGETOS
ARG TARGETARCH
ARG BUILD_SUFFIX=""
ARG BUILD_LIST="${TARGETOS}_${TARGETARCH}"
ARG SEALIGHTS_TOKEN

# Avoid safe directory git failures building with default user from go-toolset
USER root

WORKDIR /build

# Copy just the mod file for better layer caching when building locally
COPY go.mod go.sum ./
RUN go mod download

# Copy the tools/kubectl mod file for better layer caching when building locally
COPY tools/kubectl/go.mod tools/kubectl/go.sum ./tools/kubectl/
RUN cd tools/kubectl && go mod download

# Now copy everything including .git
COPY . .

# Grab the Sealight agent and instrument the Go compiler
WORKDIR /sealights
RUN \
    echo "[Sealights] Downloading Sealights Golang & CLI Agents..." && \
    wget -nv -O sealights-go-agent.tar.gz https://agents.sealights.co/slgoagent/latest/slgoagent-linux-amd64.tar.gz &&\
    wget -nv -O sealights-slcli.tar.gz https://agents.sealights.co/slcli/latest/slcli-linux-amd64.tar.gz &&\
    tar -xzf ./sealights-go-agent.tar.gz && tar -xzf ./sealights-slcli.tar.gz &&\
    rm -f ./sealights-go-agent.tar.gz ./sealights-slcli.tar.gz &&\
    ./slgoagent -v 2> /dev/null | grep version && ./slcli -v 2> /dev/null | grep version

RUN \
    echo "Initializing Sealight config with a token" && \
    ./slcli config init --lang go --token ${SEALIGHTS_TOKEN} && \
    GIT_COMMIT_SHORT="$(git -C /build rev-parse --short HEAD)" && \
    echo "Git commit short: ${GIT_COMMIT_SHORT}" && \
    echo "Creating a build session ID" && \
    ./slcli config create-bsid --app ec-cli-lucarval --branch main --build "${GIT_COMMIT_SHORT}.2" && \
    echo "Scanning the build" && \
    ./slcli scan --bsid buildSessionId.txt --path-to-scanner ./slgoagent --workspacepath /build --scm git


# Back to building
WORKDIR /build

RUN /build/build.sh "${BUILD_LIST}" "${BUILD_SUFFIX}"

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.5@sha256:d85040b6e3ed3628a89683f51a38c709185efc3fb552db2ad1b9180f2a6c38be

ARG TARGETOS
ARG TARGETARCH

LABEL \
  name="ec-cli" \
  description="Enterprise Contract verifies and checks supply chain artifacts to ensure they meet security and business policies." \
  io.k8s.description="Enterprise Contract verifies and checks supply chain artifacts to ensure they meet security and business policies." \
  summary="Provides the binaries for downloading the EC CLI. Also used as a Tekton task runner image for EC tasks. Upstream build." \
  io.k8s.display-name="Enterprise Contract" \
  io.openshift.tags="enterprise-contract ec opa cosign sigstore"

# Install tools we want to use in the Tekton task
RUN microdnf upgrade --assumeyes --nodocs --setopt=keepcache=0 --refresh && microdnf -y --nodocs --setopt=keepcache=0 install git-core jq

# Copy all the binaries so they're available to extract and download
# (Beware if you're testing this locally it will copy everything from
# your dist directory, not just the freshly built binaries.)
COPY --from=build /build/dist/ec* /usr/local/bin/

# Gzip them because that's what the cli downloader image expects, see
# https://github.com/securesign/cosign/blob/main/Dockerfile.client-server-re.rh
RUN gzip /usr/local/bin/ec_*

# Copy the one ec binary that can run in this container
COPY --from=build "/build/dist/ec_${TARGETOS}_${TARGETARCH}" /usr/local/bin/ec

# Copy the one kubectl binary that can run in this container
COPY --from=build "/build/dist/kubectl_${TARGETOS}_${TARGETARCH}" /usr/local/bin/kubectl

# Copt reduce-snapshot script needed for single component mode
COPY hack/reduce-snapshot.sh /usr/local/bin

# OpenShift preflight check requires a license
COPY --from=build /build/LICENSE /licenses/LICENSE

# OpenShift preflight check requires a non-root user
USER 1001

# Show some version numbers for troubleshooting purposes
RUN git version && jq --version && ec version && ls -l /usr/local/bin

ENTRYPOINT ["/usr/local/bin/ec"]
