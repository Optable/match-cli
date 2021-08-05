#!/usr/bin/env bash
set -eux

CLI="match-cli"

PLATFORMS="darwin/amd64"
PLATFORMS="$PLATFORMS windows/amd64"
PLATFORMS="$PLATFORMS linux/amd64"

send_file() {
  local local_path="$1"
  local remote_path="$2"
  gsutil -h "x-goog-meta-optable-cli-version:$version" cp "$local_path" "$remote_path"
}

get_version() {
  local remote_path="$1"
  gsutil ls -L "$remote_path" 2>/dev/null | grep optable-cli-version | cut -d ':' -f2
}

publish() {
  local bucket_uri=${1%/}
  local file="$2"
  local version="$3"
  local filebase; filebase=$(basename "$file")
  local prerel; prerel=$(semver get prerel "$version")

  # Send fully qualified version
  send_file "$file" "$bucket_uri/$version/$filebase" "$version"

  if [[ "$prerel" != "" ]]; then
    return 0
  fi

  # Evaluate each potential expansions
  local expansions=(
    "$bucket_uri/latest/$filebase"
  )

  local path_version
  for path in "${expansions[@]}"; do
    path_version=$(get_version "$path")
    if [[ "$path_version" == "" || "$(semver compare "$path_version" "$version")" -lt 1 ]]; then
      send_file "$file" "$path"
    fi
  done

  return 0
}

bin_filename() {
  local CLI="$1"
  local GOOS="$2"
  local GOARCH="$3"
  local BIN_FILENAME="${CLI}-${GOOS}-${GOARCH}"
  if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
  echo -n "${BIN_FILENAME}"
}

build() {
  for PLATFORM in $PLATFORMS; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    make "clean"
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} make "bin/${CLI}"
    cp "bin/${CLI}" "${1}/$(bin_filename "$CLI" "$GOOS" "$GOARCH")"
  done
}

if [ "$1" = "--build" ]; then
  build "$2"
  exit 0
fi

bucket_uri=${2%/}
version=$3

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  publish "$bucket_uri" "$(bin_filename "$CLI" "$GOOS" "$GOARCH")" "$version"
done
