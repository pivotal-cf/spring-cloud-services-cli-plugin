#!/bin/bash

set -euo pipefail

# Environment variables
readonly BUILT_PLUGIN_OUTPUT="${BUILT_PLUGIN_OUTPUT?:Environment variable BUILT_PLUGIN_OUTPUT must be set}"
readonly CF_CLI_PLUGIN_INPUT="${CF_CLI_PLUGIN_INPUT?:Environment variable CF_CLI_PLUGIN_INPUT must be set}"
readonly PLUGIN_NAME="${PLUGIN_NAME?:Environment variable PLUGIN_NAME must be set}"
readonly VERSION_INPUT="${VERSION_INPUT?:Environment variable VERSION_FILE must be set}"

readonly BUILD_VERSION="$(cat "$VERSION_INPUT/version")"

readonly GOPATH
export GOFLAGS="-mod=vendor"

run_tests() {
  go get github.com/onsi/ginkgo/ginkgo

  pushd "${CF_CLI_PLUGIN_INPUT}" > /dev/null
    "${GOPATH}"/bin/ginkgo -r
  popd > /dev/null
}

build() {
  local platform architecture binary_name version_flag
  platform="$1"
  architecture="$2"
  binary_name="${PLUGIN_NAME}-${platform}-${architecture}-${BUILD_VERSION}"
  version_flag="-X main.pluginVersion=${BUILD_VERSION}"

  if [ "${platform}" == "windows" ]; then
    binary_name="${binary_name}.exe"
  fi

  pushd "${CF_CLI_PLUGIN_INPUT}" > /dev/null
    GOOS="${platform}" GOARCH="${architecture}" go build -ldflags="${version_flag}" -o "${binary_name}"
    echo "Built ${binary_name}"
  popd > /dev/null

  mv "${CF_CLI_PLUGIN_INPUT}/${binary_name}" "${BUILT_PLUGIN_OUTPUT}/"
}

run_tests
build darwin amd64
build linux '386'
build linux amd64
build windows '386'
build windows amd64
