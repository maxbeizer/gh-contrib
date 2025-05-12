#!/bin/bash
set -e

# Set CGO_ENABLED based on target platform
build_for_platform() {
  local goos="$1"
  local goarch="$2"
  local output="$3"
  local cgo_enabled=0  # Default to disabled

  # Enable CGO only for platforms where we have the toolchain properly set up
  if [ "$goos" = "linux" ] && [ "$goarch" = "amd64" ]; then
    cgo_enabled=1
  elif [ "$goos" = "darwin" ] && ([ "$goarch" = "amd64" ] || [ "$goarch" = "arm64" ]); then
    cgo_enabled=1
  fi

  echo "Building for $goos-$goarch with CGO_ENABLED=$cgo_enabled"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED="$cgo_enabled" go build -trimpath -ldflags="-s -w" -o "$output"
}

# Create dist directory if it doesn't exist
mkdir -p dist

# Build for standard platforms
build_for_platform "darwin" "amd64" "dist/gh-contrib_${1}_darwin-amd64"
build_for_platform "darwin" "arm64" "dist/gh-contrib_${1}_darwin-arm64"
build_for_platform "linux" "386" "dist/gh-contrib_${1}_linux-386"
build_for_platform "linux" "amd64" "dist/gh-contrib_${1}_linux-amd64" 
build_for_platform "linux" "arm" "dist/gh-contrib_${1}_linux-arm"
build_for_platform "linux" "arm64" "dist/gh-contrib_${1}_linux-arm64"
build_for_platform "windows" "386" "dist/gh-contrib_${1}_windows-386.exe"
build_for_platform "windows" "amd64" "dist/gh-contrib_${1}_windows-amd64.exe"
build_for_platform "windows" "arm64" "dist/gh-contrib_${1}_windows-arm64.exe"

# Optional: Add FreeBSD if needed
build_for_platform "freebsd" "386" "dist/gh-contrib_${1}_freebsd-386"
build_for_platform "freebsd" "amd64" "dist/gh-contrib_${1}_freebsd-amd64"
build_for_platform "freebsd" "arm64" "dist/gh-contrib_${1}_freebsd-arm64"
