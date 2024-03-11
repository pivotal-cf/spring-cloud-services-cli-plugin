#!/bin/bash

set -euo pipefail

# Environment variables
readonly CF_CLI_PLUGIN_INPUT="${CF_CLI_PLUGIN_INPUT?:Environment variable CF_CLI_PLUGIN_INPUT must be set}"
readonly BUILT_PLUGIN_OUTPUT="${BUILT_PLUGIN_OUTPUT?:Environment variable BUILT_PLUGIN_OUTPUT must be set}"
readonly VERSION="${VERSION?:Environment variable VERSION must be set}"

readonly PLUGIN_NAME="spring-cloud-services-cli-plugin"
readonly PLUGIN_VERSION="${VERSION#v}"

export GOFLAGS="-mod=vendor"

function run_tests() {
  pushd "${CF_CLI_PLUGIN_INPUT}" >/dev/null

  go install github.com/onsi/ginkgo/v2/ginkgo
  ginkgo -r

  popd >/dev/null
}

function build() {
  local platform architecture binary_name version_flag
  platform="$1"
  architecture="$2"
  binary_name="${PLUGIN_NAME}-${platform}-${architecture}-${PLUGIN_VERSION}"
  version_flag="-X main.pluginVersion=${PLUGIN_VERSION}"

  if [ "${platform}" == "windows" ]; then
    binary_name="${binary_name}.exe"
  fi

  pushd "${CF_CLI_PLUGIN_INPUT}" >/dev/null
  GOOS="${platform}" GOARCH="${architecture}" CGO_ENABLED="0" go build -ldflags="${version_flag}" -o "${binary_name}"
  echo "Built ${binary_name}"
  popd >/dev/null

  mv "${CF_CLI_PLUGIN_INPUT}/${binary_name}" "${BUILT_PLUGIN_OUTPUT}/"
}

function main() {
  apt-get --allow-releaseinfo-change update && apt upgrade -y

  run_tests

  build darwin amd64
  build darwin arm64
  build linux '386'
  build linux amd64
  build windows '386'
  build windows amd64
}

main "$@"
