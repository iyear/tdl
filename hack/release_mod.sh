#!/usr/bin/env bash

# Examples:

# Update all dependencies to version vX.Y.Z
# ./hack/release_mod.sh deps vX.Y.Z

# Add tags to all modules with version vX.Y.Z
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

function deps() {
  log_info "Updating dependencies to version ${version}"
  log_info ""

  for dir in ${dirs}; do
    (
      log_info "Processing ${dir}"

      cd "${dir}"

      go mod tidy

      modules=$(go list -f '{{if not .Main}}{{if not .Indirect}}{{.Path}}{{end}}{{end}}' -m all)
      deps=$(echo "${modules}" | grep -E "${ROOT_MODULE}/.*" || true)

      for dep in ${deps}; do
        go mod edit -require "${dep}@${version}"
      done

      go mod tidy

      cd "${ROOT_DIR}"

      log_succ "  Processed ${dir}"
    )
  done

  log_succ ""
  log_succ "Dependencies updated, and commit them manually"
}

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

      tag="${prefix}${version}"
      git tag "${tag}"

      log_succ "  Tag ${tag}"
    )
  done

  log_succ ""
  log_succ "Tags added, and push them manually"
}

# run the function
"${cmd}"
