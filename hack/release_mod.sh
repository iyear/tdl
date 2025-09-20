#!/usr/bin/env bash

# Examples:

# Add tags to all submodules with version vX.Y.Z
# ./hack/release_mod.sh tags v0.1.0

set -euo pipefail

source ./hack/lib.sh

cmd="${1}"
version="${2}"
if [ -z "${version}" ]; then
  log_error "version argument is required"
  exit 2
fi

go mod tidy

dirs=$(find . -name "go.mod" -not -path "./docs/*" -exec dirname {} \;)

function tags(){
  log_info "Adding tags to all modules with version ${version}"
  log_info ""

  for dir in ${dirs}; do
    (
      log_info "Processing ${dir}"
      prefix="${dir#./}"
      prefix="${prefix#.}"
      # if prefix is not empty, append a slash
      if [ -n "${prefix}" ]; then
        prefix="${prefix}/"
      fi

      # if prefix is empty, it means it's just the root module, do not handle it
      if [ -n "${prefix}" ]; then
        tag="${prefix}${version}"
        git tag "${tag}"

        log_succ "  Tag ${tag}"
      fi
    )
  done

  log_succ ""
  log_succ "Tags added, and push them manually"
  log_succ "Then tag main module with ${version}, and push it to trigger the release"
}

# run the function
"${cmd}"
